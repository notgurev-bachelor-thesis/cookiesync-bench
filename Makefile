.PHONY: remote

remote:
	ssh root@spb-w3-stathandler.moevideo.net

deploy:
	scp -r * root@spb-w3-stathandler.moevideo.net:~/bench
