for ((i=0;i<10000;i++)) do kubectl exec -it itest -n zx -- ./itest run -c /etc/itest/itest.json c_case -n=10000; sleep 10; done

