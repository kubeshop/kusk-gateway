docker_build('kusk-gateway', '.', dockerfile='./build/manager/Dockerfile-debug')
k8s_yaml(kustomize('./config/debug'))
# k8s_resource('kusk-gateway-manager', port_forwards='40000')