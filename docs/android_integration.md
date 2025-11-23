# Android Integration Guide

This guide explains how to integrate the backend's Protocol Buffers definitions into your Android project.

## Overview

Since the backend and mobile app are in separate repositories, the recommended way to share `.proto` files is using **Git Submodules**. This ensures your Android app always uses the exact version of the API definitions that the backend is using.

## Step 1: Add Backend as Submodule

In your Android project's root directory, run:

```bash
git submodule add git@github.com:datapeice/astolfosplayer-backend.git backend
git submodule update --init --recursive
```

This will clone the backend repository into a `backend/` folder inside your Android project.

## Step 2: Configure `build.gradle.kts`

Update your `app/build.gradle.kts` to point to the proto files in the submodule.

### 1. Add Proto Source Set (The "Input")

Replace or update your `sourceSets` configuration to include the path to the protos in the submodule:

```kotlin
android {
    // ...
    sourceSets {
        getByName("main") {
            proto {
                // Point to the proto files in the submodule
                // Adjust the path if you cloned it to a different folder
                srcDir("backend/protos/proto")
            }
        }
    }
}
```

### 2. Configure Protobuf Plugin

Ensure you have the protobuf plugin configured to generate the code.

```kotlin
protobuf {
    protoc {
        artifact = "com.google.protobuf:protoc:3.24.3" // Use appropriate version
    }
    plugins {
        id("grpc") {
            artifact = "io.grpc:protoc-gen-grpc-java:1.58.0" // Use appropriate version
        }
        id("grpckt") {
            artifact = "io.grpc:protoc-gen-grpc-kotlin:1.4.0:jdk8@jar"
        }
    }
    generateProtoTasks {
        all().forEach { task ->
            task.plugins {
                id("grpc")
                id("grpckt")
            }
            task.builtins {
                id("java") {
                    option("lite")
                }
                id("kotlin")
            }
        }
    }
}
```

### 3. Register Generated Sources (The "Output")

The snippet you previously had is for registering the **generated** code so Android Studio can see it. You should keep this (or ensure your plugin version handles it automatically).

```kotlin
// Register generated code as source
sourceSets.getByName("main") {
    java.srcDir("build/generated/source/proto/main/java")
    java.srcDir("build/generated/source/proto/main/kotlin")
    java.srcDir("build/generated/source/proto/main/grpc")
    java.srcDir("build/generated/source/proto/main/grpckt")
}
```

## Summary

1.  **Input**: `sourceSets { main { proto { srcDir("backend/protos/proto") } } }` tells Gradle where to *find* the `.proto` files.
2.  **Output**: `sourceSets { main { java.srcDir("build/generated/...") } }` tells Android Studio where the *compiled* Java/Kotlin files are.
