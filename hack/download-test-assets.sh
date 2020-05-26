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
os="$(uname -s)"
os_lowercase="$(echo "$os" | tr '[:upper:]' '[:lower:]' )"

dest_dir="${1:-"${test_framework_dir}/assets/bin"}"
etcd_dest="${dest_dir}/etcd"
kubectl_dest="${dest_dir}/kubectl"
kube_apiserver_dest="${dest_dir}/kube-apiserver"
kind_dest="${dest_dir}/kind"
kustomize_dest="${dest_dir}/kustomize"

echo "About to download a couple of binaries. This might take a while..."

mkdir -p "$dest_dir"

k8s_version=1.16.4
curl $quiet -L "https://go.kubebuilder.io/test-tools/${k8s_version}/${os_lowercase}/amd64" | tar --strip-components=2 -xz -C "$dest_dir" kubebuilder/bin

kind_version=v0.8.1
curl $quiet -L "https://github.com/kubernetes-sigs/kind/releases/download/${kind_version}/kind-${os_lowercase}-amd64" --output "$kind_dest"

kustomize_version=v3.1.0
curl $quiet -L "https://github.com/kubernetes-sigs/kustomize/releases/download/${kustomize_version}/kustomize_$(echo -n $kustomize_version | sed 's/^v//')_${os_lowercase}_amd64" --output "$kustomize_dest"

chmod +x "$etcd_dest" "$kubectl_dest" "$kube_apiserver_dest" "$kind_dest" "$kustomize_dest"

echo    "# destination:"
echo    "#   ${dest_dir}"
echo    "# versions:"
echo -n "#   etcd:            "; "$etcd_dest" --version | head -n 1
echo -n "#   kube-apiserver:  "; "$kube_apiserver_dest" --version
echo -n "#   kubectl:         "; "$kubectl_dest" version --client --short
echo -n "#   kind:            "; "$kind_dest" --version
echo -n "#   kustomize:       "; "$kustomize_dest" version
