FROM jenkins/jenkins:2.479.3-lts-jdk17

ENV JAVA_OPTS="-Djenkins.install.runSetupWizard=false"

RUN jenkins-plugin-cli --plugins configuration-as-code structs

COPY jenkins-step-rpc-plugin.hpi /usr/share/jenkins/ref/plugins/
COPY casc.yaml /opt/casc.yaml

ENV CASC_JENKINS_CONFIG=/opt/casc.yaml
