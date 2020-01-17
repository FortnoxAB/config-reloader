library 'whatsout'

node('go1.13') {
	container('run'){
		def newTag = ''
		def tag = ''
		def gitTag = ''

		try {
			stage('Checkout'){
				checkout scm
				notifyBitbucket()
				gitTag = sh(script: 'git tag -l --contains HEAD', returnStdout: true).trim()
			}

			stage('Fetch dependencies'){
				sh('go mod download')
			}
			stage('Run test'){
				sh('make test')
			}

			if(gitTag != ''){
				tag = gitTag
			}

			if( tag != ''){
				strippedTag = tag.replaceFirst('v', '')
				stage('Build the application'){
					echo "Building with docker tag ${strippedTag}"
					sh('CGO_ENABLED=0 GOOS=linux go build')
				}

				stage('Generate docker image'){
					image = docker.build('fortnox/config-reloader:'+strippedTag, '--pull .')
				}

				stage('Push docker image'){
					docker.withRegistry("https://quay.io", 'docker-registry') {
						image.push()
					}
				}
			}
			currentBuild.result = 'SUCCESS'
		} catch(err) {
			currentBuild.result = 'FAILED'
			notifyBitbucket()
			throw err
		}
		notifyBitbucket()
	}
}

