export CNI_COMMAND=ADD
export CNI_CONTAINERID=f81d4fae-7dec-11d0-a765-00a0c91e6bf6
export CNI_NETNS="my-ip-netns"
export CNI_IFNAME="my-ifname"
export CNI_PATH="$PATH:/home/ubuntu/go/src/github.com/appc/cni/plugins/ipam/host-local"
#./cni-infoblox < run.conf
./cni-infoblox < ipam.conf
