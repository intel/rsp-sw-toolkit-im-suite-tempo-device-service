rrpBuildGoCode {
    projectKey = 'tempo-device-service'
    testDependencies = ['mongo']
    dockerBuildOptions = ['--squash', '--build-arg GIT_COMMIT=$GIT_COMMIT']
    ecrRegistry = "280211473891.dkr.ecr.us-west-2.amazonaws.com"
    dockerImageName = "rsp/${projectKey}"
    protexProjectName = 'bb-tempo-device-service'
    buildImage = 'amr-registry.caas.intel.com/rrp/ci-go-build-image:1.12.0-alpine'

    infra = [
        stackName: 'RSP-Codepipeline-TempoDeviceService'
    ]

    customBuildScript = "make build"

    notify = [
        slack: [ success: '#ima-build-success', failure: '#ima-build-failed' ]
    ]
}
