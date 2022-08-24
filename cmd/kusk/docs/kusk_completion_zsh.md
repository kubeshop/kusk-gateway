## kusk completion zsh

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(kusk completion zsh); compdef _kusk kusk

To load completions for every new session, execute once:

#### Linux:

	kusk completion zsh > "${fpath[1]}/_kusk"

#### macOS:

	kusk completion zsh > $(brew --prefix)/share/zsh/site-functions/_kusk

You will need to start a new shell for this setup to take effect.


```
kusk completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.kusk.yaml)
```

### SEE ALSO

* [kusk completion](kusk_completion.md)	 - Generate the autocompletion script for the specified shell

