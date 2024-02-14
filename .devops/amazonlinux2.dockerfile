# Will be working only with cgroups v1
# How to check?
# if [ -e /sys/fs/cgroup/cgroup.controllers ]; then echo "Using cgroup v2"; else echo "Using cgroup v1"; fi
# Which OS works on cgroups v1?
# i.e. Ubuntu 20.04 LTS

FROM amazonlinux:2
RUN yum update -y
RUN yum install curl git vim-enhanced systemd systemd-sysv net-tools iputils initscripts -y
ENV container docker
RUN cd /lib/systemd/system/sysinit.target.wants/ ; \
        for i in *; do [ $i = systemd-tmpfiles-setup.service ] || rm -f $i ; done ; \
        rm -f /lib/systemd/system/multi-user.target.wants/* ; \
        rm -f /etc/systemd/system/*.wants/* ; \
        rm -f /lib/systemd/system/local-fs.target.wants/* ; \
        rm -f /lib/systemd/system/sockets.target.wants/*udev* ; \
        rm -f /lib/systemd/system/sockets.target.wants/*initctl* ; \
        rm -f /lib/systemd/system/basic.target.wants/* ; \
        rm -f /lib/systemd/system/anaconda.target.wants/*

RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmybash/oh-my-bash/master/tools/install.sh)"
#
#RUN yum install -y python3 python3-pip wget tar make cmake gcc  ; \
#    yum groupinstall -y "Development Tools" ; \
#    yum install -y libcap-devel gperf glib2-devel libmount-devel ; \
#    pip3 install meson ; \
#    pip3 install ninja ; \
#    pip3 install jinja2
#
#RUN wget https://github.com/systemd/systemd/archive/refs/tags/v249.tar.gz ; \
#    tar -xf v249.tar.gz ; \
#    cd systemd-249 ; \
#    ./configure ; \
#    make ; \
#    /usr/local/bin/ninja -C build install

VOLUME ["/sys/fs/cgroup"]
CMD ["/sbin/init"]
#CMD ["/usr/lib/systemd/systemd", "--system"]