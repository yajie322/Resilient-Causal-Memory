all:
	docker build -t rcmi .
run:
	docker run -d --name node1 -ip 172.17.0.1 rcmi
	docker run -d --name node2 -ip 172.17.0.2 rcmi
	docker run -d --name node3 -ip 172.17.0.3 rcmi
	docker run -d --name node4 -ip 172.17.0.4 rcmi