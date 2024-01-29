package cmd

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"hutao/pkg/graceful"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	polling int
	domains string
	wg      sync.WaitGroup
)

var RootCmd = &cobra.Command{
	Use:   "hutao",
	Short: "tls/ssl expiration check and notic",
}

// 使用deployment方式部署，不够灵活，轮询时间以部署的时间每24小时一次
var pollingJobCmd = &cobra.Command{
	Use:     "polling",
	Aliases: []string{"images-sync"},
	Short:   "tls/ssl expiration check and notic",
	Long:    `Automatically detect certificate expiration time and trigger message reminders！by fermi`,
	RunE: func(cmd *cobra.Command, args []string) error {
		pollingTime := time.Duration(polling) * time.Second
		log.Log().Msgf("Hutao polling begining:%v :)", pollingTime)
		domainsSlice := strings.Split(domains, ",")
		//  优雅轮询并且启动健康检查，并且在接收到失败信号好，结束程序
		graceful.NeverStopByTicker(":8000", time.NewTicker(pollingTime), func() {
			log.Info().Msg("Hutao normal operation, bro～")
			Checks(domainsSlice)
		})

		log.Log().Msg("Hutao polling shutting down :)")

		return nil
	},
}

// 使用cronjob的方式，这里执行一次性脚本，轮询时间由k8s cronjob来控制！
var cronjobCmd = &cobra.Command{
	Use:     "cronjob",
	Aliases: []string{"images-sync-crobjob"},
	Short:   "tls/ssl expiration check and notic",
	Long:    `Automatically detect certificate expiration time and trigger message reminders！by fermi`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Log().Msgf("Hutao cronjob begining:%v :)", polling)
		domainsSlice := strings.Split(domains, ",")
		Checks(domainsSlice)
		log.Log().Msg("Hutao cronjob shutting down :)")
		return nil
	},
}

func init() {
	// shortvideo-api-dev.lingverse.co,shortvideo-api.lingverse.co 已迁移到k8s的cert-manager
	RootCmd.PersistentFlags().StringVar(&domains, "domains", "www.baidu.com,www.google.com", "需要被检查的域名列表,用逗号隔开")
	RootCmd.PersistentFlags().IntVarP(&polling, "polling", "o", 86400, "轮询检查的时间间隔，默认24小时执行一次")
	RootCmd.AddCommand(pollingJobCmd, cronjobCmd)
}

// Execute executes the RootCmd
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func Checks(domains []string) {
	port := "443"
	warnDays := 15
	certList := map[string]string{}

	wg.Add(len(domains))
	for _, domain := range domains {
		domain := domain
		go func() {
			defer wg.Done()
			expiryTime, err := checkSSLExpiry(domain, port, warnDays)
			if err != nil {
				log.Err(err).Msgf("Failed to check SSL expiry for %v:%v.\n expiryTime: %v", domain, port, expiryTime.String())
				return
			}
			certList[domain] = expiryTime.String()
			log.Log().Msgf("SSL certificate for %v:%v will expire at %v", domain, port, expiryTime.String())
		}()
	}

	wg.Wait()

	jsonStr, _ := json.MarshalIndent(certList, "", "  ")
	log.Log().Msgf("Success certList is %v", string(jsonStr))
}

func checkSSLExpiry(domain string, port string, warnDays int) (time.Time, error) {
	dialer := &net.Dialer{
		Timeout: time.Second * 5,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", net.JoinHostPort(domain, port), nil)
	if err != nil {
		return time.Time{}, err
	}
	defer func(conn *tls.Conn) {
		err := conn.Close()
		if err != nil {
			log.Err(err).Msg("tcp close fail")
		}
	}(conn)

	cert := conn.ConnectionState().PeerCertificates[0]

	// Check if the certificate is expired
	if time.Now().AddDate(0, 0, warnDays).After(cert.NotAfter) {
		return cert.NotAfter, fmt.Errorf("cert is going to expire in less than %d days", warnDays)
	}

	return cert.NotAfter, nil
}
