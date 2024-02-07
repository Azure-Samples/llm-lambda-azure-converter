# AWS Lambda To Azure Function Converter

The objective of this project is to create a cli app that is able to convert AWS Lambda Functions to Azure Functions, with the help of Azure Open AI.

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

## Resources

Developed using as reference the LATS implementation by Andy Zhou, Kai Yan, Michal Shlapentokh-Rothman, Haohan Wang and Yu-Xiong Wang

@misc{zhou2023language,
      title={Language Agent Tree Search Unifies Reasoning Acting and Planning in Language Models}, 
      author={Andy Zhou and Kai Yan and Michal Shlapentokh-Rothman and Haohan Wang and Yu-Xiong Wang},
      year={2023},
      eprint={2310.04406},
      archivePrefix={arXiv},
      primaryClass={cs.AI}
}

To understand the implementation please give a look to the following [article](./article/Article.md).





