package runtimeknobs

import (
	"net/netip"
	"os"
	"strings"
	"time"

	"github.com/example/appfoundrylab/backend/pkg/env"
)

const (
	RequestLogTrustedProxyCIDRsEnv            = "REQUEST_LOG_TRUSTED_PROXY_CIDRS"
	DefaultLoggerHealthTimeout                = 1500 * time.Millisecond
	DefaultLoggerIngestTimestampMaxAge        = 5 * time.Minute
	DefaultLoggerIngestTimestampMaxFutureSkew = 5 * time.Second
)

func LoggerHealthTimeout() time.Duration {
	return positiveDurationFromEnv("LOGGER_HEALTH_TIMEOUT_MS", DefaultLoggerHealthTimeout, time.Millisecond)
}

func LoggerIngestTimestampMaxAge() time.Duration {
	return positiveDurationFromEnv(
		"LOGGER_INGEST_TIMESTAMP_MAX_AGE_SECONDS",
		DefaultLoggerIngestTimestampMaxAge,
		time.Second,
	)
}

func LoggerIngestTimestampMaxFutureSkew() time.Duration {
	return positiveDurationFromEnv(
		"LOGGER_INGEST_TIMESTAMP_MAX_FUTURE_SKEW_SECONDS",
		DefaultLoggerIngestTimestampMaxFutureSkew,
		time.Second,
	)
}

func RequestLogTrustedProxyCIDRs() []string {
	prefixes, _ := ParseTrustedProxyPrefixes(os.Getenv(RequestLogTrustedProxyCIDRsEnv))
	return trustedProxyCIDRs(prefixes)
}

func ParseTrustedProxyPrefixes(raw string) ([]netip.Prefix, []string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	prefixes := make([]netip.Prefix, 0, len(parts))
	rejected := make([]string, 0)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		prefix, err := parseTrustedProxyPrefix(part)
		if err != nil {
			rejected = append(rejected, part)
			continue
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes, rejected
}

func positiveDurationFromEnv(name string, defaultValue time.Duration, unit time.Duration) time.Duration {
	raw := env.GetIntWithDefault(name, int(defaultValue/unit))
	if raw <= 0 {
		return defaultValue
	}
	return time.Duration(raw) * unit
}

func parseTrustedProxyPrefix(raw string) (netip.Prefix, error) {
	if strings.Contains(raw, "/") {
		return netip.ParsePrefix(raw)
	}

	addr, err := netip.ParseAddr(raw)
	if err != nil {
		return netip.Prefix{}, err
	}
	return netip.PrefixFrom(addr, addr.BitLen()), nil
}

func trustedProxyCIDRs(prefixes []netip.Prefix) []string {
	if len(prefixes) == 0 {
		return nil
	}

	values := make([]string, 0, len(prefixes))
	for _, prefix := range prefixes {
		values = append(values, prefix.String())
	}
	return values
}
