# micro-app
## Introduction
This is a simple micro application framework that mainly supports the creation of components through configuration files. 
The plug-in component development mode can make our code clearer and have lower coupling.

It is suitable for a simple single process application, although it also supports multiple processes, 
it does not achieve multi process communication. If your application requires multiple processes on a single machine, 
then this framework will not be suitable for your application.

## Directory Structure Description
N/A

## Configuration Template Description
This is a framework configuration file for simulating an application, which displays the framework configuration structure.
This application has an application ID of "SimApp" and a log level of DEBUG.  
'gc_control' means to allow the framework to intervene in GC operations. You can selectively turn it on or off, but we generally do not turn it on.  
'configs' means the required configuration file path, which can monitor file changes and generate update events for upper level application.  
'sub_process_list' means a list of sub processes that need to be started, which requires a boot file path and log prefix. 
I do not recommend using this framework to handle sub processes.  
'components' means a list of components that define the components that need to be run and the parameters required by the components.
```json
{
  "app_id": "SimApp",
  "log_level": "DEBUG",
  "gc_control": {
    "disable_default_gc": false,
    "enable_force": false,
    "force_policy": {
      "interval_seconds": 120,
      "mem_peak": 4096
    }
  },
  "configs": [{
    "key": "http_api_routes",
    "path": "/etc/config/simapp/http_api_routes.json"
  },
    {
      "key": "static_resource_hash_list",
      "path": "/etc/config/simapp/static_resource_hash_list.json"
    }
  ],
  "sub_process_list": {
    "enable": false,
    "commands": [{
      "launcher_cfg": "/etc/config/simapp/sub_launcher.json",
      "log_out_prefix": "SimAppSub1"
    }]
  },
  "components": [{
    "component_type": "HTTPAPIServer",
    "disable": false,
    "kw": {
      "server_addr": "0.0.0.0:8080",
      "handler_workers_num": 100
    }
  },
    {
      "component_type": "StaticResourceServer",
      "disable": false,
      "kw": {
        "server_addr": "0.0.0.0:8081",
        "access_key": "b4a3e6dc20eaa6d"
      }
    }
  ]
}
```

## Example
Please refer to the directory path 'micro-app/example'