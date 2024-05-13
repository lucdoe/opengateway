workspace {

    model {
        user = person "Developer" {
            description "Configures services and plugins through a YAML file."
        }

        softwareSystem = softwareSystem "OpenGateway" {
            description "Routes requests to various services based on the YAML configurations."
        }

        client = softwareSystem "Client" {
            description "Requests the services through the API Gateway."
            tags "External"
        }

        service = softwareSystem "Services" {
            description "Service that is registered and routed through the API Gateway."
            tags "External"
        }

        user -> softwareSystem "Configures API Gateway via YAML"
        client -> softwareSystem "Sends requests to API Gateway"
        softwareSystem -> service "Reverse proxies requests to services"
    }

    views {
        systemContext softwareSystem {
            include *
            autolayout lr
        }

        styles {
            element "External" {
                background #D3D3D3 
                color #000000     
            }
        }

        theme default
    }
}
