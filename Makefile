all:
	docker build -t rcmi .

run:
	docker run -d --name svr1 -e id=1 rcmi
	docker run -d --name svr2 -e id=2 rcmi
	docker run -d --name svr3 -e id=3 rcmi
	docker run -d --name svr4 -e id=4 rcmi

client:
	docker run -it --name client -e type=c rcmi

rm:
	docker rm svr1
	docker rm svr2
	docker rm svr3
	docker rm svr4
	docker rm client
