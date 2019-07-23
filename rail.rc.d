#!/bin/sh

# PROVIDE: rail
# REQUIRE: LOGIN
# KEYWORD: shutdown
#
# Add the following lines to /etc/rc.conf.local or /etc/rc.conf
# to enable this service:
#
# rail_enable (bool)
#     Set to NO by default
#     Set it to YES to enable rail
# rail_user (string)
#     Set user to run rail
#     Default is "rail"
# rail_group (string)
#     Set group to run rail
#     Default is "rail"
# rail_args (string)
#     Set additional command line arguments
#     Default is ""

. /etc/rc.subr

name=rail
rcvar=rail_enable

load_rc_config $name

: ${rail_enable:="NO"}
: ${rail_user:="daemon"}
: ${rail_group:="daemon"}
: ${rail_args:=""}

pidfile="/var/run/${name}.pid"
required_files="${rail_config}"
command="/usr/sbin/daemon"
procname="/usr/local/bin/${name}"
sig_reload="TERM"
extra_commands="reload"
command_args="-p ${pidfile} -m 3 -T ${name} \
                /usr/bin/env ${procname} \
                ${rail_args}"

start_precmd=rail_startprecmd

rail_startprecmd()
{
    if [ ! -e ${pidfile} ]; then
        install \
            -o ${rail_user} \
            -g ${rail_group} \
            /dev/null ${pidfile};
    fi
}

load_rc_config $name
run_rc_command "$1"
