package api

import (
	"encoding/json"
	"net/http"

	"github.com/Ullaakut/nmap/v3"
	"github.com/gin-gonic/gin"
	"github.com/vpaklatzis/conduit-go/models"
)

var req models.ScanParamsRequest

func (s *Server) ScanHandler(c *gin.Context) {
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	n, err := configureScanner(c)
	if err != nil {
		s.log.Fatalf("Failed to create the nmap scanner: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	done := make(chan error)
	result, warnings, err := n.Async(done).Run()
	if err != nil {
		s.log.Fatalf("Unable to run nmap scan: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	// Blocks main until the scan has completed.
	if err := <-done; err != nil {
		// Warnings are non-critical errors from nmap.
		if len(*warnings) > 0 {
			s.log.Infof("Run finished with warnings: %s\n", *warnings)
		}
		s.log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	for _, host := range result.Hosts {
		if len(host.Ports) == 0 {
			s.log.Fatal("No target hosts were found.")
			continue
		}
		if len(host.Addresses) == 0 {
			s.log.Fatal("No addresses of the provided target hosts were found.")
			continue
		}
		//fmt.Printf("Host %q:\n", host.Addresses[0])
		p := &models.Port{}
		var ports []models.Port

		for _, port := range host.Ports {
			//fmt.Printf("\tPort %d/%s %s %s\n", port.ID, port.Protocol, port.State.State, port.Service.Name)
			p.Port = port.ID
			p.Protocol = port.Protocol
			p.State = port.State.State
			p.ServiceName = port.Service.Name
			p.ServiceVersion = port.Service.Version

			ports = append(ports, *p)
		}
		response := &models.ScanParamsResponse{
			Host:  result.Hosts[0].Addresses[0].Addr,
			Ports: ports,
		}
		r, err := json.Marshal(response)
		if err != nil {
			s.log.Fatal(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		c.Data(http.StatusOK, "application/json", r)
	}
}

func configureScanner(c *gin.Context) (*nmap.Scanner, error) {
	n, err := nmap.NewScanner(
		c,
		nmap.WithTargets(req.Target),
		nmap.WithPorts(req.Port),
		nmap.WithServiceInfo(),
		//nmap.WithOSDetection(),
	)
	return n, err
}
