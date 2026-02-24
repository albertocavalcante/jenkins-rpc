plugins {
    id("org.jetbrains.kotlin.jvm") version "2.2.21"
    id("org.jenkins-ci.jpi") version "0.55.1"
}

repositories {
    mavenCentral()
    maven("https://repo.jenkins-ci.org/public/")
}

jenkinsPlugin {
    jenkinsVersion = providers.gradleProperty("jenkinsVersion").get()
    shortName = "jenkins-step-rpc-plugin"
    displayName = "Jenkins Step RPC Plugin"
    url = "https://github.com/albertocavalcante/jenkins-rpc"
    gitHubUrl = "https://github.com/albertocavalcante/jenkins-rpc"
}

dependencies {
    implementation("org.jetbrains.kotlin:kotlin-stdlib:2.2.21")
    implementation("com.google.protobuf:protobuf-java-util:4.32.1")
    testImplementation("org.jetbrains.kotlin:kotlin-test-junit5:2.2.21")
    testImplementation("org.junit.vintage:junit-vintage-engine:5.11.4")
}

sourceSets {
    named("main") {
        java.srcDir("../contracts/gen/java")
        java.srcDir(layout.buildDirectory.dir("generated-src/localizer"))
    }
}

tasks.withType<org.jetbrains.kotlin.gradle.tasks.KotlinCompile>().configureEach {
    dependsOn("localizeMessages")
    compilerOptions {
        jvmTarget.set(org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_17)
    }
}

tasks.test {
    useJUnitPlatform()
}
