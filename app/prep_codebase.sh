# Tidy and format the code
go mod tidy && go fmt ./... 

# Generate the swagger documentation
swag fmt
swag init -g ./src/app/bootstrap/bootstrap.go -o ./docs --parseDependency