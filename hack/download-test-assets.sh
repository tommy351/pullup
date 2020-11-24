#!/usr/bin/env bash
set -eu

# Use DEBUG=1 ./scripts/download-binaries.sh to get debug output
quiet="-s"
[[ -z "${DEBUG:-""}" ]] || {
  set -x
  quiet=""
}

logEnd() {
  local msg='done.'
  [ "$1" -eq 0 ] || msg='Error downloading assets'
  echo "$msg"
}
trap 'logEnd $?' EXIT

test_framework_dir="$(cd "$(dirname "$0")/.." ; pwd)"
goos=$(go env GOOS)
goarch=$(go env GOARCH)

dest_dir="${1:-"${test_framework_dir}/assets/bin"}"
etcd_dest="${dest_dir}/etcd"
kubectl_dest="${dest_dir}/kubectl"
kube_apiserver_dest="${dest_dir}/kube-apiserver"
kind_dest="${dest_dir}/kind"

echo "About to download a couple of binaries. This might take a while..."

mkdir -p "$dest_dir"

k8s_version=1.19.2
curl $quiet -L "https://go.kubebuilder.io/test-tools/${k8s_version}/${goos}/${goarch}" | tar --strip-components=2 -xz -C "$dest_dir" kubebuilder/bin

kind_version=v0.9.0
curl $quiet -L "https://github.com/kubernetes-sigs/kind/releases/download/${kind_version}/kind-${goos}-${goarch}" --output "$kind_dest"

chmod +x "$etcd_dest" "$kubectl_dest" "$kube_apiserver_dest" "$kind_dest"

echo    "# destination:"
echo    "#   ${dest_dir}"
echo    "# versions:"
echo -n "#   etcd:            "; "$etcd_dest" --version | head -n 1
echo -n "#   kube-apiserver:  "; "$kube_apiserver_dest" --version
echo -n "#   kubectl:         "; "$kubectl_dest" version --client --short
echo -n "#   kind:            "; "$kind_dest" --version
