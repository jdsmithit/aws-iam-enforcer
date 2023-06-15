# AWS-IAM-ENFORCER

## Introduction

The AWS IAM Enforcer is a powerful tool designed to enforce and maintain consistent Identity and Access Management (IAM) policies across your AWS resources. It helps you ensure that only authorized actions can be performed and provides an extra layer of security for your AWS environment.

## Features
Disable IAM User Credentials: With the AWS IAM Enforcer, you can automate disabling IAM user credentials when the reach a certain age.
Disable Unused IAM User Credentials: With the AWS IAM Enforcer, you can automate the disabling of IAM user credentials when they reach a certain age and are also not used.

## Usage
Deploy the terraform to the AWS accounts

### Lambda Enviroment Variables

DRY_RUN_MODE = true/false = default(false)
DISABLE_KEYS_TOGGLE = true/false = default(false)
DISABLE_KEYS_DAYS = Any Number = default(30)
DISABLE_UNUSED_ONLY_TOGGLE = true/false = default(false)
DISABLE_UNUSED_ONLY_KEYS_DAYS = Any Number = default(30)

