FROM ubuntu

ADD infoblox-daemon /usr/local/bin/infoblox-daemon


ENTRYPOINT ["/usr/local/bin/infoblox-daemon"]
