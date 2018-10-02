all:
	docker build -t rcmi .
run:
	docker run -d --name node1 -e id=0 -p 8888:8888 rcmi
	docker run -d --name node2 -e id=1 -p 8889:8889 rcmi
	docker run -d --name node3 -e id=2 -p 8890:8890 rcmi
	docker run -d --name node4 -e id=3 -p 8891:8891 rcmi