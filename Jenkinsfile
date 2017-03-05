  env.DOCKERHUB_USERNAME = 'vipconsult'

  node("docker-test") {
    checkout scm

    stage("Unit Test") {
      sh "docker run --rm -v ${WORKSPACE}:/go/src/cd-demo golang go test cd-demo -v --run Unit"
    }
    stage("Integration Test") {
      try {
        sh "docker build -t cd-demo ."
        sh "docker rm -f cd-demo || true"
        sh "docker run -d -p 8080:8080 --name=cd-demo cd-demo"
        // env variable is used to set the server where go test will connect to run the test
        sh "docker run --rm -v ${WORKSPACE}:/go/src/cd-demo --link=cd-demo -e SERVER=cd-demo golang go test cd-demo -v --run Integration"
      }
      catch(e) {
        error "Integration Test failed"
      }finally {
        sh "docker rm -f cd-demo || true"
        sh "docker ps -aq | xargs docker rm || true"
        sh "docker images -aq -f dangling=true | xargs docker rmi || true"
      }
    }
    stage("Build") {
      sh "docker build -t ${DOCKERHUB_USERNAME}/cd-demo:${BUILD_NUMBER} ."
    }
    stage("Publish") {
      withDockerRegistry([credentialsId: 'DockerHub']) {
        sh "docker push ${DOCKERHUB_USERNAME}/cd-demo:${BUILD_NUMBER}"
      }
    }
  }

  node("docker-stage") {
    checkout scm

    stage("Staging") {
      try {
        sh "docker rm -f cd-demo || true"
        sh "docker run -d -p 8080:8080 --name=cd-demo ${DOCKERHUB_USERNAME}/cd-demo:${BUILD_NUMBER}"
        sh "docker run --rm -v ${WORKSPACE}:/go/src/cd-demo --link=cd-demo -e SERVER=cd-demo golang go test cd-demo -v"

      } catch(e) {
        error "Staging failed"
      } finally {
        sh "docker rm -f cd-demo || true"
        sh "docker ps -aq | xargs docker rm || true"
        sh "docker images -aq -f dangling=true | xargs docker rmi || true"
      }
    }
  }

  node("docker-prod") {
    stage("Production") {
      try {
        // Create the service if it doesn't exist otherwise just update the image
        sh '''
          SERVICES=$(docker service ls --filter name=cd-demo --quiet | wc -l)
          if [[ "$SERVICES" -eq 0 ]]; then
            docker network rm cd-demo || true
            docker network create --driver overlay --attachable cd-demo
            docker service create --replicas 3 --network cd-demo --name cd-demo -p 8080:8080 ${DOCKERHUB_USERNAME}/cd-demo:${BUILD_NUMBER}
          else
            docker service update --image ${DOCKERHUB_USERNAME}/cd-demo:${BUILD_NUMBER} cd-demo
          fi
          '''
        // run some final tests in production
        checkout scm
        sh '''
          sleep 60s 
          for i in `seq 1 20`;
          do
            STATUS=$(docker service inspect --format '{{ .UpdateStatus.State }}' cd-demo)
            if [[ "$STATUS" != "updating" ]]; then
              docker run --rm -v ${WORKSPACE}:/go/src/cd-demo --network cd-demo -e SERVER=cd-demo golang go test cd-demo -v --run Integration
              break
            fi
            sleep 10s
          done
          
        '''
      }catch(e) {
        sh "docker service update --rollback  cd-demo"
        error "Service update failed in production"
      }finally {
        sh "docker ps -aq | xargs docker rm || true"
      }
    }
  }
