# 🔐 HTTPS Setup Instructions

## ✅ Що було налаштовано:

### 1. **Backend API**: `https://api.example.com`
- Самопідписаний SSL сертифікат створений
- OAuth redirect URI оновлений на `https://api.example.com/auth/callback`
- CORS налаштований для HTTPS доменів

### 2. **Frontend**: `https://app.example.com`
- API конфігурація оновлена для використання `https://api.example.com`
- Самопідписаний SSL сертифікат створений
- Ingress налаштований для обох доменів

## 🚀 Як тестувати:

### 1. **Додайте домени в /etc/hosts** (вже зроблено):
```bash
echo "192.168.49.2 api.example.com app.example.com" | sudo tee -a /etc/hosts
```

### 2. **Тестування API**:
```bash
# Health check
curl -k https://api.example.com/health

# Auth endpoints
curl -k https://api.example.com/auth/login
```

### 3. **Тестування Frontend**:
Відкрийте у браузері: `https://app.example.com`
(Браузер покаже попередження про самопідписаний сертифікат - натисніть "Продовжити")

## ⚠️ Важливо для Google OAuth:

Оновіть Google OAuth Console:
1. Перейдіть до [Google Cloud Console](https://console.cloud.google.com/)
2. Виберіть ваш проект
3. Перейдіть до APIs & Services > Credentials
4. Знайдіть ваш OAuth 2.0 Client ID
5. **Додайте новий Authorized redirect URI**:
   ```
   https://api.example.com/auth/callback
   ```
6. Збережіть зміни

## 🔄 Для продакшену з реальним доменом:

1. **Купіть справжній домен** (наприклад: `yourapp.com`)
2. **Налаштуйте DNS** A record на IP вашого Kubernetes кластера
3. **Оновіть Ingress**:
   ```yaml
   # Замініть example.com на ваш реальний домен
   - host: api.yourapp.com
   - host: app.yourapp.com
   ```
4. **Увімкніть Let's Encrypt**:
   ```yaml
   annotations:
     cert-manager.io/cluster-issuer: "letsencrypt-prod"
   ```
5. **Оновіть Google OAuth Console** з реальними URL

## 🛠️ Команди для перевірки:

```bash
# Статус Ingress
kubectl get ingress

# Статус сертифікатів
kubectl get secrets | grep tls

# Логи pods
kubectl logs -l app=go-api
kubectl logs -l app=react-frontend

# Перезапуск deployments
kubectl rollout restart deployment/go-api
kubectl rollout restart deployment/react-frontend
```

## 🎯 Поточний стан:
- ✅ Backend API: `https://api.example.com` (працює)
- ✅ Frontend: `https://app.example.com` (готовий)
- ✅ SSL сертифікати: самопідписані (для локального тестування)
- ⚠️ Google OAuth: потрібно оновити redirect URI в консолі
