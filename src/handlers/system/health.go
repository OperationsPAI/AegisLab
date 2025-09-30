package system

import (
	"aegis/client"
	"aegis/client/k8s"
	"aegis/config"
	"aegis/database"
	"aegis/dto"
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetHealth handles system health check
//
//	@Summary System health check
//	@Description Get system health status and service information
//	@Tags System
//	@Produce json
//	@Success 200 {object} dto.GenericResponse[dto.HealthCheckResponse] "Health check successful"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /system/health [get]
func GetHealth(c *gin.Context) {
	start := time.Now()

	services := make(map[string]dto.ServiceInfo)
	overallStatus := "healthy"

	buildkitInfo := checkBuildKitHealth()
	services["buildkit"] = buildkitInfo
	if buildkitInfo.Status != "healthy" {
		overallStatus = "unhealthy"
	}

	dbInfo := checkDatabaseHealth()
	services["database"] = dbInfo
	if dbInfo.Status != "healthy" {
		overallStatus = "unhealthy"
	}

	jaegerInfo := checkJaegerHealth()
	services["jaeger"] = jaegerInfo
	if jaegerInfo.Status != "healthy" {
		overallStatus = "unhealthy"
	}

	k8sInfo := checkKubernetesHealth()
	services["kubernetes"] = k8sInfo
	if k8sInfo.Status != "healthy" {
		overallStatus = "unhealthy"
	}

	redisInfo := checkRedisHealth()
	services["redis"] = redisInfo
	if redisInfo.Status != "healthy" {
		overallStatus = "unhealthy"
	}

	response := dto.HealthCheckResponse{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Uptime:    time.Since(start).String(),
		Services:  services,
	}

	dto.SuccessResponse(c, response)
}

// checkBuildKitHealth checks BuildKit daemon connectivity
func checkBuildKitHealth() dto.ServiceInfo {
	start := time.Now()

	buildkitAddr := config.GetString("buildkit.address")

	conn, err := net.DialTimeout("tcp", buildkitAddr, 5*time.Second)
	responseTime := time.Since(start)

	if err != nil {
		return dto.ServiceInfo{
			Status:       "unhealthy",
			LastChecked:  time.Now(),
			ResponseTime: responseTime.String(),
			Error:        "BuildKit daemon unreachable",
			Details:      fmt.Sprintf("Cannot connect to BuildKit at %s: %v", buildkitAddr, err),
		}
	}
	defer conn.Close()

	return dto.ServiceInfo{
		Status:       "healthy",
		LastChecked:  time.Now(),
		ResponseTime: responseTime.String(),
	}
}

// checkDatabaseHealth checks database connectivity
func checkDatabaseHealth() dto.ServiceInfo {
	start := time.Now()

	if database.DB == nil {
		return dto.ServiceInfo{
			Status:       "unhealthy",
			LastChecked:  time.Now(),
			ResponseTime: "N/A",
			Error:        "Database connection not available",
		}
	}

	// Test connection with a simple query
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result int
	err := database.DB.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error
	responseTime := time.Since(start)

	if err != nil {
		return dto.ServiceInfo{
			Status:       "unhealthy",
			LastChecked:  time.Now(),
			ResponseTime: responseTime.String(),
			Error:        "Database query failed",
			Details:      err.Error(),
		}
	}

	return dto.ServiceInfo{
		Status:       "healthy",
		LastChecked:  time.Now(),
		ResponseTime: responseTime.String(),
	}
}

// checkJaegerHealth checks Jaeger tracing service connectivity
func checkJaegerHealth() dto.ServiceInfo {
	start := time.Now()

	jaegerURL := fmt.Sprintf("http://%s/v1/traces", config.GetString("jaeger.endpoint"))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "HEAD", jaegerURL, nil)
	if err != nil {
		return dto.ServiceInfo{
			Status:       "unhealthy",
			LastChecked:  time.Now(),
			ResponseTime: time.Since(start).String(),
			Error:        "Failed to create Jaeger OTLP request",
			Details:      err.Error(),
		}
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	responseTime := time.Since(start)

	if err != nil {
		return dto.ServiceInfo{
			Status:       "unhealthy",
			LastChecked:  time.Now(),
			ResponseTime: responseTime.String(),
			Error:        "Jaeger OTLP endpoint unreachable",
			Details:      err.Error(),
		}
	}
	defer resp.Body.Close()

	// OTLP endpoints typically return 405 Method Not Allowed for HEAD requests
	if resp.StatusCode != http.StatusMethodNotAllowed && resp.StatusCode != http.StatusOK {
		return dto.ServiceInfo{
			Status:       "unhealthy",
			LastChecked:  time.Now(),
			ResponseTime: responseTime.String(),
			Error:        fmt.Sprintf("Jaeger OTLP returned unexpected status %d", resp.StatusCode),
		}
	}

	return dto.ServiceInfo{
		Status:       "healthy",
		LastChecked:  time.Now(),
		ResponseTime: responseTime.String(),
		Details:      "Jaeger OTLP endpoint responding",
	}
}

// checkKubernetesHealth checks Kubernetes API connectivity
func checkKubernetesHealth() dto.ServiceInfo {
	start := time.Now()

	// Try to get Kubernetes config
	restConfig := k8s.GetK8sRestConfig()
	if restConfig == nil {
		return dto.ServiceInfo{
			Status:       "unhealthy",
			LastChecked:  time.Now(),
			ResponseTime: time.Since(start).String(),
			Error:        "Kubernetes config not available",
		}
	}

	// Create Kubernetes client
	k8sClient := k8s.GetK8sClient()
	if k8sClient == nil {
		return dto.ServiceInfo{
			Status:       "unhealthy",
			LastChecked:  time.Now(),
			ResponseTime: time.Since(start).String(),
			Error:        "Kubernetes client not available",
		}
	}

	// Create Kubernetes dynamic client
	k8sDynamicClient := k8s.GetK8sDynamicClient()
	if k8sDynamicClient == nil {
		return dto.ServiceInfo{
			Status:       "unhealthy",
			LastChecked:  time.Now(),
			ResponseTime: time.Since(start).String(),
			Error:        "Kubernetes dynamic client not available",
		}
	}

	// Test API connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := k8sClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return dto.ServiceInfo{
			Status:       "unhealthy",
			LastChecked:  time.Now(),
			ResponseTime: time.Since(start).String(),
			Error:        "Kubernetes API request failed",
			Details:      err.Error(),
		}
	}

	return dto.ServiceInfo{
		Status:       "healthy",
		LastChecked:  time.Now(),
		ResponseTime: time.Since(start).String(),
	}
}

// checkRedisHealth checks Redis connectivity
func checkRedisHealth() dto.ServiceInfo {
	start := time.Now()

	rdb := client.GetRedisClient()
	if rdb == nil {
		return dto.ServiceInfo{
			Status:       "unhealthy",
			LastChecked:  time.Now(),
			ResponseTime: "N/A",
			Error:        "Redis connection not available",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test connection with PING
	result := rdb.Ping(ctx)
	responseTime := time.Since(start)

	if result.Err() != nil {
		return dto.ServiceInfo{
			Status:       "unhealthy",
			LastChecked:  time.Now(),
			ResponseTime: responseTime.String(),
			Error:        result.Err().Error(),
		}
	}

	return dto.ServiceInfo{
		Status:       "healthy",
		LastChecked:  time.Now(),
		ResponseTime: responseTime.String(),
	}
}
