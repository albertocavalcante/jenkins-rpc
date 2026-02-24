module github.com/albertocavalcante/jenkins-rpc/tests/e2e

go 1.26.0

require (
	github.com/albertocavalcante/jenkins-rpc/contracts v0.0.0
	github.com/albertocavalcante/jenkins-rpc/go-client v0.0.0
	google.golang.org/protobuf v1.36.10
)

replace (
	github.com/albertocavalcante/jenkins-rpc/contracts => ../../contracts
	github.com/albertocavalcante/jenkins-rpc/go-client => ../../go-client
)
