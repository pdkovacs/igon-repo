#!/bin/awk -f

function checkColumns()
{
    if ($2 != "TYPE") {
        print "Unexpedted 2th column: " $2 ". (Expected TYPE)" | "cat 1>&2"
        exit 1
    }
    if ($4 != "EXTERNAL-IP") {
        print "Unexpedted 4th column: " $4 ". (Expected EXTERNAL-IP)" | "cat 1>&2"
        exit 1
    }
}

BEGIN {
	print "cat /etc/hosts > tmp-hosts;" | "/bin/bash"

    nr = 0
	while (("kubectl get svc" | getline) > 0) {
        if (nr == 0) {
            checkColumns()
            nr++;
        }
		if ($2 ~ /LoadBalancer/) {
			printf("grep -v %s tmp-hosts > tmp-hosts-tmp && mv tmp-hosts-tmp tmp-hosts;", $1) | "/bin/bash";
			printf("echo %s\t%s >> tmp-hosts;", $4, $1)  | "/bin/bash";
		}
	}
	close("kubectl get svc")
	print "cat tmp-hosts; rm tmp-hosts" | "/bin/bash"
    close("/bin/bash")
}
