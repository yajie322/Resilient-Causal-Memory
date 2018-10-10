all:
	docker build -t rcmi .
run:
	docker run -d --name svr1 -e id=0 rcmi
	docker run -d --name svr2 -e id=1 rcmi
	docker run -d --name svr3 -e id=2 rcmi
	docker run -d --name svr4 -e id=3 rcmi

client:
	docker run -it --name client -e type=c rcmi

run1:
	docker run -it --name svr1 -e id=0 -p 8888:8888 rcmi

run2:
	docker run -it --name svr2 -e id=1 -p 8889:8889 rcmi

run3:
	docker run -it --name svr3 -e id=2 -p 8890:8890 rcmi

run4:
	docker run -it --name svr4 -e id=3 -p 8891:8891 rcmi

stop:
	docker stop svr1
	docker stop svr2
	docker stop svr3
	docker stop svr4

rm:
	docker rm svr1
	docker rm svr2
	docker rm svr3
	docker rm svr4
	docker rm client