package main

import (
    "fmt"
    "net/http"
    "sync"
)

// On passe le WaitGroup sous forme de pointeur pour partager le même compteur
func scannerCible(url string, wg *sync.WaitGroup) {
    defer wg.Done()

    resp, err := http.Get(url)
    if err != nil {
        fmt.Printf("[ERREUR] %s : Erreur de connexion\n", url)
        return
    }
    defer resp.Body.Close()
    
    fmt.Printf("[SUCCÈS] %s : Statut HTTP %d\n", url, resp.StatusCode)
}

func main() {
    cibles := []string{
        "http://127.0.0.1",
        "https://google.com",
        "https://github.com",
        "http://une-url-qui-n-existe-pas.lol",
    }

    fmt.Println("Début du scan multi-cibles...")

    var wg sync.WaitGroup

    for _, cible := range cibles {
        wg.Add(1)
        
        go scannerCible(cible, &wg)
    }

    wg.Wait()
    
    fmt.Println("Scan terminé !")
}