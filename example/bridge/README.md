This folder have example of how to bring up 3 node kubernetes cluster using
bridge network type and infoblox plugin.

Troubleshooting step for kube-dns failure
-----------------------------------------

If kube-dns gets IP from the cni-infoblox-daemon's default network(172.18.0.0/16) as follows
cmd:kubectl get pods -o wide --all-namespaces

```
kube-system kube-dns-6f4fd4bdf-hb68k 1/3 CrashLoopBackOff 19018 21d 172.18.0.2 master
```

Once CNI network conf is copied to all the nodes then delete the kube-dns so that proper IP will get assigned as follows 
cmd:"kubectl delete pod kube-dns-6f4fd4bdf-hb68k -n kube-system"

```
kube-system kube-dns-6f4fd4bdf-6ddvp 3/3 Running 0 25s 10.15.20.3 master
kube-system kube-dns-6f4fd4bdf-hb68k 0/3 Terminating 19018 21d 172.18.0.2 master
```

