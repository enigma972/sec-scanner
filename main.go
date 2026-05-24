package main

import (
    "fmt"
    "net/http"
    "sync"
    "encoding/json"
)

type RapportScan struct {
    Cible   string `json:"target_url"`
    Statut  int    `json:"status_code"`
    EnLigne bool   `json:"is_online"`
}

// On passe le WaitGroup sous forme de pointeur pour partager le même compteur
func scannerCible(url string, wg *sync.WaitGroup, ch chan<- RapportScan) {
    defer wg.Done()

    resp, err := http.Get(url)
    if err != nil {
        fmt.Printf("[ERREUR] %s : Erreur de connexion\n", url)
        ch <- RapportScan{Cible: url, Statut: 0, EnLigne: false}
        return
    }
    defer resp.Body.Close()
    
    fmt.Printf("[SUCCÈS] %s : Statut HTTP %d\n", url, resp.StatusCode)
    ch <- RapportScan{Cible: url, Statut: resp.StatusCode, EnLigne: resp.StatusCode == http.StatusOK}
}

func handlerScan(w http.ResponseWriter, r *http.Request) {
    defer r.Body.Close()

    queryParams := r.URL.Query()

    if !queryParams.Has("urls") {
        http.Error(w, "Url parameter missing", http.StatusBadRequest)
        return
    }

    urls := queryParams["urls"]

    fmt.Println("Début du scan multi-cibles...")

    var wg sync.WaitGroup
    resultsCh := make(chan RapportScan, len(urls))

    for _, cible := range urls {
        wg.Add(1)
        go scannerCible(cible, &wg, resultsCh)
    }

    // Ferme le channel une fois toutes les goroutines terminées
    go func() {
        wg.Wait()
        close(resultsCh)
    }()

    // Récupère tous les rapports depuis le channel
    var rapports []RapportScan
    for r := range resultsCh {
        rapports = append(rapports, r)
    }

    fmt.Println("Scan terminé !")

    // On renvoie la liste complète des rapports en JSON
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(rapports)
}

func main() {
    // On associe la route "/api/scan" à notre fonction handlerScan
    http.HandleFunc("/api/scan", handlerScan)

    fmt.Println("Serveur démarré sur le port 8080...")

    // C'est ici que tu interviens ! 👇
    http.ListenAndServe(":80", nil)
}