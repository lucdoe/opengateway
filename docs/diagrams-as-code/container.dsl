workspace {

    model {
        user = person "Developer" {
            description "Configures services and plugins through a YAML file."
        }
        
        client = softwareSystem "Client" {
            description "Requests the services through the API Gateway."
            tags "External"
        }
        
        service = softwareSystem "Services" {
            description "Service that is registered and routed through the API Gateway."
            tags "External"
        }
        
        softwareSystem = softwareSystem "OpenGateway" {
            configLoader = container "Configuration Loader" {
                description "Loads and parses service configurations from YAML file."
                technology "Go"
            }

            requestHandler = container "Request Handlers" {
                description "Processes incoming requests and applies plugins as middleware."
                technology "Go"
            }

            middleware = container "Middleware" {
                description "Apply plugins as middleware to requests and responses."
                technology "Go"
            }

            proxyRouter = container "Proxy Router" {
                description "Proxies requests to the configured services and back."
                technology "Go"
            }

            yamlConfig = container "YAML Configuration" {
                description "Stores service configurations and plugin settings in a YAML file format."
                technology "File"
                tags "Data Store"
            }
            
            user -> yamlConfig "Edits"
            client -> requestHandler "Sends requests to API Gateway"
            configLoader -> yamlConfig "Reads and parses"
            configLoader -> requestHandler "Provides configurations to"
            requestHandler -> middleware "Applies"
            middleware -> proxyRouter "Routes through"
            proxyRouter -> service "Forwards request and gets response"
        }
    }

    views {
        container softwareSystem {
            include *
            autolayout lr
            
            }
        styles {
                element "Data Store" {
                    shape "Cylinder"
                }
                element "External" {
                background #D3D3D3 
                color #000000     
            }
        }

        theme default
    }
}
