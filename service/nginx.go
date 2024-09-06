package service

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

type NginxService interface {
	// RunApp(conf models.Config) error
	UpdateOrInstallNginx() error
	AddServer(serverName, serverPort, protocol string) error
	CheckIfServerExists(serverName, serverPort, protocol string) (bool, error)
	ResetNginxToDefault() error
}

type nginxSerivce struct {
	dockerClient DockerClient
}

func NewNginxSerivce(dockerClient DockerClient) NginxService {
	return &nginxSerivce{
		dockerClient: dockerClient,
	}
}

func (n *nginxSerivce) CheckIfServerExists(serverName, serverPort, protocol string) (bool, error) {
	nginxConf := fmt.Sprintf("/etc/nginx/sites-available/%s", serverName)

	// Check if a configuration for the same server name already exists
	if _, err := os.Stat(nginxConf); err == nil {

		confFile, err := os.Open(nginxConf)
		if err != nil {
			return false, fmt.Errorf("failed to open Nginx configuration file: %v", err)
		}
		defer confFile.Close()

		// Read the existing configuration file content
		confData := make([]byte, 1024)
		_, err = confFile.Read(confData)
		if err != nil {
			return false, fmt.Errorf("failed to read Nginx configuration file: %v", err)
		}

		// Use the new function to generate the required configuration
		generatedConf, err := prepareNginxConf(serverName, serverPort, protocol)
		if err != nil {
			return false, err
		}

		// Compare the existing configuration with the generated configuration
		if strings.Contains(string(confData), generatedConf) {
			// The configuration already exists
			return true, nil
		}
	}

	// Configuration doesn't exist
	return false, nil
}

func (n *nginxSerivce) AddServer(serverName, serverPort, protocol string) error {
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

	// Use the new function to generate the Nginx configuration
	confContent, err := prepareNginxConf(serverName, serverPort, protocol)
	if err != nil {
		return err
	}

	// Create the new Nginx configuration file
	confFile, err := os.Create(nginxConf)
	if err != nil {
		return fmt.Errorf("failed to create Nginx configuration file: %v", err)
	}
	defer confFile.Close()

	// Write the configuration content to the file
	_, err = confFile.WriteString(confContent)
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

func (n *nginxSerivce) UpdateOrInstallNginx() error {
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

func (n *nginxSerivce) ResetNginxToDefault() error {
	fmt.Println("Resetting NGINX to its default state...")

	// Stop NGINX service
	fmt.Println("Stopping NGINX...")
	if err := exec.Command("sudo", "systemctl", "stop", "nginx").Run(); err != nil {
		return fmt.Errorf("failed to stop NGINX: %v", err)
	}

	// Remove custom configurations
	fmt.Println("Removing custom configurations in /etc/nginx/sites-available and /etc/nginx/sites-enabled...")
	if err := os.RemoveAll("/etc/nginx/sites-available"); err != nil {
		return fmt.Errorf("failed to remove sites-available directory: %v", err)
	}
	if err := os.RemoveAll("/etc/nginx/sites-enabled"); err != nil {
		return fmt.Errorf("failed to remove sites-enabled directory: %v", err)
	}

	// Reinstall NGINX to reset to its default state
	fmt.Println("Reinstalling NGINX...")
	if err := exec.Command("sudo", "apt", "update").Run(); err != nil {
		return fmt.Errorf("failed to update package list: %v", err)
	}
	if err := exec.Command("sudo", "apt", "install", "--reinstall", "-y", "nginx").Run(); err != nil {
		return fmt.Errorf("failed to reinstall NGINX: %v", err)
	}

	// Restore default nginx.conf
	fmt.Println("Restoring default nginx.conf...")
	if err := exec.Command("sudo", "cp", "/etc/nginx/nginx.conf.default", "/etc/nginx/nginx.conf").Run(); err != nil {
		return fmt.Errorf("failed to restore default nginx.conf: %v", err)
	}

	// Start NGINX
	fmt.Println("Starting NGINX...")
	if err := exec.Command("sudo", "systemctl", "start", "nginx").Run(); err != nil {
		return fmt.Errorf("failed to start NGINX: %v", err)
	}

	fmt.Println("NGINX has been reset to its default state.")
	return nil
}

func prepareNginxConf(serverName, serverPort, protocol string) (string, error) {
	// Nginx configuration template
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

	// Create a template instance
	tmpl, err := template.New("nginx").Parse(confTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse Nginx template: %v", err)
	}

	// Prepare the template with the actual values
	var generatedConf bytes.Buffer
	err = tmpl.Execute(&generatedConf, struct {
		ServerName string
		ServerPort string
		Protocol   string
	}{
		ServerName: serverName,
		ServerPort: serverPort,
		Protocol:   protocol,
	})
	if err != nil {
		return "", fmt.Errorf("failed to execute Nginx template: %v", err)
	}

	// Return the generated configuration as a string
	return generatedConf.String(), nil
}
