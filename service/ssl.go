package service

import (
	"fmt"
	"os/exec"
)

type SSLService interface {
	SetupSSL(domain, email string) error
}

type sslService struct{}

func NewSSLService() SSLService {
	return &sslService{}
}

func (s *sslService) SetupSSL(domain, email string) error {
	// Ensure domain is provided
	if domain == "" {
		return fmt.Errorf("domain name is required")
	}

	// Check if Certbot is installed
	if _, err := exec.LookPath("certbot"); err != nil {
		fmt.Println("Certbot not found, installing...")
		if err := exec.Command("sudo", "apt", "update").Run(); err != nil {
			return fmt.Errorf("failed to update package list: %v", err)
		}
		if err := exec.Command("sudo", "apt", "install", "-y", "certbot", "python3-certbot-nginx").Run(); err != nil {
			return fmt.Errorf("failed to install Certbot: %v", err)
		}
	} else {
		fmt.Println("Certbot is already installed")
	}

	// Generate SSL certificate
	fmt.Printf("Generating SSL certificate for %s...\n", domain)
	cmd := exec.Command("sudo", "certbot", "--nginx", "-d", domain, "--non-interactive", "--agree-tos", "-m", email, "--redirect")
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Println(string(output))
		return fmt.Errorf("failed to create certificate for %s: %v", domain, err)
	} else {
		fmt.Println(string(output))
		fmt.Printf("Certificate for %s created successfully\n", domain)
	}

	// Test NGINX configuration and reload
	fmt.Println("Testing Nginx configuration...")
	if err := exec.Command("sudo", "nginx", "-t").Run(); err != nil {
		return fmt.Errorf("NGINX configuration test failed: %v", err)
	}

	fmt.Println("Reloading Nginx...")
	if err := exec.Command("sudo", "systemctl", "reload", "nginx").Run(); err != nil {
		return fmt.Errorf("failed to reload Nginx: %v", err)
	}

	// Setup auto-renewal
	fmt.Println("Setting up auto-renewal...")
	if err := exec.Command("sudo", "systemctl", "enable", "certbot.timer").Run(); err != nil {
		return fmt.Errorf("failed to enable auto-renewal: %v", err)
	}
	if err := exec.Command("sudo", "systemctl", "start", "certbot.timer").Run(); err != nil {
		return fmt.Errorf("failed to start auto-renewal: %v", err)
	}

	fmt.Println("Auto-renewal setup completed")
	fmt.Printf("SSL setup complete for %s\n", domain)
	return nil
}
