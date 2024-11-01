build:
	CGO_ENABLED=0 go build -o action1 main.go
deploy:
	kubectl apply -f ./deployment.yaml
images:
	docker build -t demo1 .
web:
	kubectl apply -f ./nginx/