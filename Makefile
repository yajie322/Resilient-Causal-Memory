all:
	docker build -t rcmi .
run:
	docker run -d --name node1 -e id=0 -ip 172.17.0.1 rcmi
	docker run -d --name node2 -e id=1 -ip 172.17.0.2 rcmi
	docker run -d --name node3 -e id=2 -ip 172.17.0.3 rcmi
	docker run -d --name node4 -e id=3 -ip 172.17.0.4 rcmi