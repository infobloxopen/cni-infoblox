FROM ubuntu

ADD infoblox-cni-daemon /usr/local/bin/infoblox-cni-daemon


ENTRYPOINT ["/usr/local/bin/infoblox-cni-daemon"]
