package handlers

import (
    "expo-open-ota/internal/assets"
    cdn2 "expo-open-ota/internal/cdn"
    "expo-open-ota/internal/compression"
    "expo-open-ota/internal/metrics"
    "expo-open-ota/internal/services"
    "github.com/google/uuid"
    "log"
    "net/http"
)

func AssetsHandler(w http.ResponseWriter, r *http.Request) {
	channelName := r.Header.Get("expo-channel-name")
	preventCDNRedirection := r.Header.Get("prevent-cdn-redirection") == "true"
	branchMap, err := services.FetchExpoChannelMapping(channelName)
	if err != nil {
		log.Printf("[RequestID: %s] Error fetching channel mapping: %v", uuid.New().String(), err)
        clientId := r.Header.Get("EAS-Client-ID")
        platform := r.URL.Query().Get("platform")
        runtimeVersion := r.URL.Query().Get("runtimeVersion")
        updateId := r.Header.Get("expo-current-update-id")
        branch := ""
        if branchMap != nil {
            branch = branchMap.BranchName
        }
        metrics.TrackUpdateErrorUser(clientId, platform, runtimeVersion, branch, updateId)
		http.Error(w, "Error fetching channel mapping", http.StatusInternalServerError)
		return
	}
	if branchMap == nil {
		log.Printf("[RequestID: %s] No branch mapping found for channel: %s", uuid.New().String(), channelName)
        clientId := r.Header.Get("EAS-Client-ID")
        platform := r.URL.Query().Get("platform")
        runtimeVersion := r.URL.Query().Get("runtimeVersion")
        updateId := r.Header.Get("expo-current-update-id")
        metrics.TrackUpdateErrorUser(clientId, platform, runtimeVersion, "", updateId)
		http.Error(w, "No branch mapping found", http.StatusNotFound)
		return
	}

	req := assets.AssetsRequest{
		Branch:         branchMap.BranchName,
		AssetName:      r.URL.Query().Get("asset"),
		RuntimeVersion: r.URL.Query().Get("runtimeVersion"),
		Platform:       r.URL.Query().Get("platform"),
		RequestID:      uuid.New().String(),
	}

	cdn := cdn2.GetCDN()
	if cdn == nil || preventCDNRedirection {
		resp, err := assets.HandleAssetsWithFile(req)
		if err != nil {
            clientId := r.Header.Get("EAS-Client-ID")
            metrics.TrackUpdateErrorUser(clientId, req.Platform, req.RuntimeVersion, req.Branch, r.Header.Get("expo-current-update-id"))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		for key, value := range resp.Headers {
			w.Header().Set(key, value)
		}
		if resp.StatusCode != 200 {
            clientId := r.Header.Get("EAS-Client-ID")
            metrics.TrackUpdateErrorUser(clientId, req.Platform, req.RuntimeVersion, req.Branch, r.Header.Get("expo-current-update-id"))
			http.Error(w, string(resp.Body), resp.StatusCode)
			return
		}
		compression.ServeCompressedAsset(w, r, resp.Body, resp.ContentType, req.RequestID)
		return
	}
	resp, err := assets.HandleAssetsWithURL(req, cdn)
	if err != nil {
        clientId := r.Header.Get("EAS-Client-ID")
        metrics.TrackUpdateErrorUser(clientId, req.Platform, req.RuntimeVersion, req.Branch, r.Header.Get("expo-current-update-id"))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, resp.URL, http.StatusFound)
}
