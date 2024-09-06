package service

import (
	"fmt"
	"os"
	"os/exec"
	"text/template"
)

func manageServer(serverName, serverPort, protocol string) error {
	// Validate protocol
	if protocol != "http" && protocol != "https" {
		return fmt.Errorf("Invalid protocol specified. Use 'http' or 'https'.")
	}

	// Update the package list and install Nginx if not installed
	if err := updateAndInstallNginx(); err != nil {
		return err
	}

	// Configure Nginx
	if err := configureNginx(serverName, serverPort, protocol); err != nil {
		return err
	}

	return nil
}

func updateAndInstallNginx() error {
	fmt.Println("Updating package list...")
	if err := exec.Command("sudo", "apt", "update").Run(); err != nil {
		return fmt.Errorf("failed to update package list: %v", err)
	}

	// Check if Nginx is installed
	if _, err := exec.LookPath("nginx"); err != nil {
		fmt.Println("Nginx not found. Installing Nginx...")
		if err := exec.Command("sudo", "apt", "install", "-y", "nginx").Run(); err != nil {
			return fmt.Errorf("failed to install Nginx: %v", err)
		}
	}
	return nil
}

func configureNginx(serverName, serverPort, protocol string) error {
	nginxConf := fmt.Sprintf("/etc/nginx/sites-available/%s", serverName)

	// Check if a configuration for the same server name already exists
	if _, err := os.Stat(nginxConf); err == nil {
		fmt.Printf("Removing existing configuration for %s...\n", serverName)
		if err := os.Remove(nginxConf); err != nil {
			return fmt.Errorf("failed to remove existing configuration: %v", err)
		}
		if err := os.Remove(fmt.Sprintf("/etc/nginx/sites-enabled/%s", serverName)); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove existing symlink: %v", err)
		}
	}

	// Create new Nginx server configuration
	confTemplate := `server {
    server_name {{.ServerName}};

    location / {
        proxy_pass {{.Protocol}}://localhost:{{.ServerPort}};
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
`

	tmpl, err := template.New("nginx").Parse(confTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse Nginx template: %v", err)
	}

	confFile, err := os.Create(nginxConf)
	if err != nil {
		return fmt.Errorf("failed to create Nginx configuration file: %v", err)
	}
	defer confFile.Close()

	err = tmpl.Execute(confFile, struct {
		ServerName string
		ServerPort string
		Protocol   string
	}{
		ServerName: serverName,
		ServerPort: serverPort,
		Protocol:   protocol,
	})
	if err != nil {
		return fmt.Errorf("failed to write Nginx configuration: %v", err)
	}

	// Link the configuration to sites-enabled
	if err := os.Symlink(nginxConf, fmt.Sprintf("/etc/nginx/sites-enabled/%s", serverName)); err != nil {
		return fmt.Errorf("failed to create symlink: %v", err)
	}

	// Test the Nginx configuration for syntax errors
	fmt.Println("Testing Nginx configuration...")
	if err := exec.Command("sudo", "nginx", "-t").Run(); err != nil {
		return fmt.Errorf("Nginx configuration test failed. Please check the configuration: %v", err)
	}

	// Reload Nginx to apply the changes
	fmt.Println("Reloading Nginx...")
	if err := exec.Command("sudo", "nginx", "-s", "reload").Run(); err != nil {
		return fmt.Errorf("failed to reload Nginx: %v", err)
	}

	fmt.Printf("Nginx configuration for %s successfully reloaded.\n", serverName)
	return nil
}
