# 03 - Kotlin Feasibility

## Question

Can this Jenkins plugin be implemented in Kotlin while remaining aligned with Jenkins plugin infrastructure?

## Findings

1. Jenkins plugin artifacts are JVM plugins (`hpi`) and can be produced from Kotlin sources.
2. Jenkins ecosystem contains Kotlin-authored plugins in `jenkinsci` organization.
3. Kotlin build setups exist in both Maven (`kotlin-maven-plugin`) and Gradle (`org.jetbrains.kotlin.jvm` + `org.jenkins-ci.jpi`) plugin projects.

## Evidence To Track

1. `jenkinsci/zdevops-plugin` contains a Kotlin Maven `hpi` setup.
2. `jenkinsci/gamekins-plugin` uses Kotlin with Gradle Jenkins plugin tooling.
3. `jenkinsci/gradle-convention-plugin` documents modern Kotlin support for plugin projects.

## Decision

Proceed with Kotlin for plugin implementation, but keep architecture and API design independent from language choice.
