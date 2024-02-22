# AWS Lambda To Azure Function Converter

The objective of this project is to create a cli app that is able to convert AWS Lambda Functions to Azure Functions, with the help of Azure Open AI.

> **Disclaimer:** This repository is provided **as-is** and is not officially supported. It is an experimental implementation that explores the use of Large Language Models (LLMs) to convert AWS Lambda functions into Azure Functions. Please note that this implementation does not guarantee successful conversion and the outcome may vary depending on the specific code and LLM used. If you wish to experiment with the code, you are welcome to fork it.

## Features

This project sample provides the following features:

* Running converter from Lambda to Azure Functions
* Lambda to Azure Function examples

## Getting Started

### Prerequisites

- Go 1.20 or bigger
- An Azure OpenAI with a GPT-4 deployment 

### Quickstart

To start running the Jupyter notebook, follow the steps in the [jupyter notebook for vscode guide](https://code.visualstudio.com/docs/datascience/jupyter-notebooks) to setup your conda environment in VS Code.

Using [example config](app/example-config.yaml) as reference create a `app/config.yaml` with your api key and endpoint.

Type `Ctrl+Shift+D` or go to the `Run and Debug` tab and select `Run cli convert`.

Now you should be able to see the demo running.

## How it works 

To understand the implementation please give a look to the following [article](https://techcommunity.microsoft.com/t5/azure-architecture-blog/converting-an-aws-lambda-function-in-go-into-an-azure-function/ba-p/4054916).

## References

Developed using as reference the LATS implementation [[1]](#1).

<a id="1">[1]</a>
Andy Zhou and Kai Yan and Michal Shlapentokh-Rothman and Haohan Wang and Yu-Xiong Wang (2023)
Language Agent Tree Search Unifies Reasoning Acting and Planning in Language Models
https://arxiv.org/abs/2310.04406v2






