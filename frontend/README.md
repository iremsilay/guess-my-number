# Guess My Number Frontend

React, TypeScript ve Vite ile hazırlanmıştır.

## Çalıştırma

Önce backend'i ayrı bir terminalde başlat:

```bash
cd backend
go run .
```

Repo kökünden ikinci bir terminal aç ve frontend'i başlat:

```bash
cd frontend
npm install
npm run dev
```

Tarayıcıda `http://localhost:5173` adresini aç. Geliştirme sunucusu `/games`
isteklerini `http://localhost:8080` adresindeki backend'e yönlendirir.

## Kontroller

- Fiziksel klavyedeki `0-9` tuşları veya ekrandaki tuşlar rakam ekler.
- `Backspace` ve `Delete` son rakamı siler.
- `Enter` ve `GO` dört rakam tamamlandığında tahmini gönderir.
- Aynı rakam bir tahminde yalnızca bir kez kullanılabilir.
- Kod sıfırla başlayabilir.

## Production API adresi

Frontend ve backend farklı adreslerde yayınlanacaksa build öncesinde API adresini
tanımla:

```bash
VITE_API_URL=https://api.example.com npm run build
```

Bu durumda backend'in frontend adresine CORS izni vermesi gerekir.
