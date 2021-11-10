// package local provides implementation of the non-k8s Envoy Control Plane service that watches the
// provided OpenAPI file with x-kusk extentions and updates Envoy configuration.
package local

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/getkin/kin-openapi/openapi3"

	envoyConfig "github.com/kubeshop/kusk-gateway/envoy/config"
	envoyConfigManager "github.com/kubeshop/kusk-gateway/envoy/manager"
	"github.com/kubeshop/kusk-gateway/spec"
)

func RunLocalService(apiSpecPath string, envoyControlPlaneAddr string) {

	envoyConfigMgr := envoyConfigManager.New(context.Background(), envoyControlPlaneAddr, nil)
	go func() {
		if err := envoyConfigMgr.Start(); err != nil {
			log.Fatal(err, "unable to start Envoy xDS API Server")
			os.Exit(1)
		}
	}()

	if err := parseAndApply(apiSpecPath, envoyConfigMgr); err != nil {
		log.Println("Parsing file error: ", err)
	}
	// After that - subscribe to changes to that file, apply if modified and block
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("detected modified file:", event.Name)
					// FIXME: for some reason the file, passed to parsing is broken when modified
					// presumably IDE (VSCode) doesn't finish writing when it is picked up.
					// Remove this ugly sleep once found the way to detect finished write
					time.Sleep(time.Second)
					log.Println("parsing ", event.Name)
					if err := parseAndApply(event.Name, envoyConfigMgr); err != nil {
						log.Println("parsing file error, skipped this processing:", err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(apiSpecPath)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func parseAndApply(apiSpecPath string, envoyMgr *envoyConfigManager.EnvoyConfigManager) error {
	// parse OpenAPI spec
	apiSpec, err := spec.NewParser(openapi3.NewLoader()).Parse(apiSpecPath)
	if err != nil {
		return err
	}

	// parse x-kusk top-level extension
	kuskExtensionOpts, err := spec.GetOptions(apiSpec)
	if err != nil {
		return err
	}
	if err = kuskExtensionOpts.FillDefaultsAndValidate(); err != nil {
		return err
	}
	envoyConfig := envoyConfig.New()
	if err := envoyConfig.UpdateConfigFromAPIOpts(kuskExtensionOpts, apiSpec); err != nil {
		return err
	}
	snapshot, err := envoyConfig.GenerateSnapshot()
	if err != nil {
		return err
	}
	if err = envoyMgr.ApplyNewFleetSnapshot(envoyConfigManager.DefaultFleetName, snapshot); err != nil {
		return err
	}
	return nil
}
