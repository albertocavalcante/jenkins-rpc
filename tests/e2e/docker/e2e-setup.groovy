import jenkins.model.Jenkins
import hudson.security.AuthorizationStrategy

def instance = Jenkins.get()
instance.setAuthorizationStrategy(AuthorizationStrategy.UNSECURED)
instance.save()
