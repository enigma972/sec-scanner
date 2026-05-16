package main

import (
    "flag"
    "fmt"
	"net/http"
)

func main() {
    // Déclaration d'un flag "--target" avec une valeur par défaut et une description
    target := flag.String("target", "http://127.0.0.1", "L'adresse IP cible à scanner")
    
    // Cette fonction lit les arguments du terminal et les assigne aux variables
    flag.Parse()

    // Comment afficherais-tu la valeur de target ici ?
	fmt.Printf("Lancement du scan sur la cible : %s\n", *target)

	resp, err := http.Get(*target)

	if err != nil {

		fmt.Println("Erreur de connexion")

		return

	}

	defer resp.Body.Close()
	
	fmt.Printf("Cible en ligne! Statut HTTP : %d\n", resp.StatusCode)
}