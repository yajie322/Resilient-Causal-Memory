all:
	docker build -t rcmi .
run:
	docker run -d --name node1 -e id=0 -p 8888:8888 rcmi
	docker run -d --name node2 -e id=1 -p 8889:8889 rcmi
	docker run -d --name node3 -e id=2 -p 8890:8890 rcmi
	docker run -d --name node4 -e id=3 -p 8891:8891 rcmi

run1:
	docker run -it --name node1 -e id=0 -p 8888:8888 rcmi

run2:
	docker run -it --name node2 -e id=1 -p 8889:8889 rcmi

run3:
	docker run -it --name node3 -e id=2 -p 8890:8890 rcmi

run4:
	docker run -it --name node4 -e id=3 -p 8891:8891 rcmi

stop:
	docker stop node1
	docker stop node2
	docker stop node3
	docker stop node4

rm:
	docker rm node1
	docker rm node2
	docker rm node3
	docker rm node4