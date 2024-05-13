workspace {

    model {
    softwareSystem = softwareSystem "API Gateway" {
        container = container "OpenGateway" {
            parser = component "Configuration Parser" {
                description "Parses the YAML configuration files."
                technology "Go"
            }

            main = component "Main" {
                description "Initializes and configures the server and middleware."
                technology "Go"
            }

            server = component "Server" {
                description "Handles web server operations, routes management and chaining middleware."
                technology "Go"
            }

            proxy = component "Proxy" {
                description "Forwards requests to configured services and replies response."
                technology "Go"
            }

            middleware = component "Middleware" {
                description "Supplys middleware for the configured plugins."
                technology "Go"
            }

            plugins = component "Plugins" {
                description "Modular components such as Authentication, Rate-limiting, CORS, Caching, and Logging."
                technology "Go"
            }

            parser -> main "Provides configuration"
            main -> server "Sends configuration and initalizes"
            server -> proxy "Uses"
            proxy -> server "Sends service response"
            plugins -> middleware "Supply buisness logic"

            config = component "Config" {
                description "Contains gateway, service, and plugin configurations."
                tags "Data Store"
                technology "File"
            }

            parser -> config "Reads and parses"
            main -> middleware "Passes configurations to middleware"
            middleware -> plugins "Applies configurations to plugins"
            middleware -> server "Provide configured middleware"
        }
    }
    }

    views {
        component container {
            include *
            autolayout lr
        }
        
         styles {
                element "Data Store" {
                    shape "Cylinder"
                }
        }

        theme default
    }
}
