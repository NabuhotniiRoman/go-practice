#!/bin/bash

echo "🚀 Deploying Go API to https://api.example.com"

# Перевіряємо чи встановлений cert-manager
echo "📋 Checking cert-manager installation..."
if ! kubectl get crd certificates.cert-manager.io &> /dev/null; then
    echo "🔧 Installing cert-manager..."
    kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.12.0/cert-manager.yaml
    echo "⏳ Waiting for cert-manager to be ready..."
    kubectl wait --for=condition=ready pod -l app=cert-manager -n cert-manager --timeout=300s
fi

# Перевіряємо чи встановлений nginx-ingress
echo "📋 Checking nginx-ingress installation..."
if ! kubectl get ns ingress-nginx &> /dev/null; then
    echo "🔧 Installing nginx-ingress..."
    kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.0/deploy/static/provider/cloud/deploy.yaml
    echo "⏳ Waiting for nginx-ingress to be ready..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/component=controller -n ingress-nginx --timeout=300s
fi

# Застосовуємо конфігурації
echo "📝 Applying Kubernetes configurations..."

# Замініть email в cluster-issuer.yaml на ваш реальний email
read -p "📧 Enter your email for Let's Encrypt certificates: " EMAIL
sed -i "s/your-email@example.com/$EMAIL/g" /home/rn/My\ Go\ app/go-practice/k8s/cluster-issuer.yaml

kubectl apply -f /home/rn/My\ Go\ app/go-practice/k8s/cluster-issuer.yaml
kubectl apply -f /home/rn/My\ Go\ app/go-practice/k8s/configmap.yaml
kubectl apply -f /home/rn/My\ Go\ app/go-practice/k8s/service.yaml
kubectl apply -f /home/rn/My\ Go\ app/go-practice/k8s/go-api-deployment.yaml
kubectl apply -f /home/rn/My\ Go\ app/go-practice/k8s/ingress.yaml

echo "⏳ Waiting for deployment to be ready..."
kubectl rollout status deployment/go-api --timeout=300s

echo "🔍 Checking ingress status..."
kubectl get ingress go-app-ingress

echo "📋 Getting external IP..."
kubectl get svc -n ingress-nginx

echo "✅ Deployment completed!"
echo ""
echo "📝 Next steps:"
echo "1. Point your DNS A record for api.example.com to the external IP above"
echo "2. Wait a few minutes for DNS propagation"
echo "3. Update Google OAuth Console with new redirect URI: https://api.example.com/auth/callback"
echo "4. Test your API at: https://api.example.com/health"
echo ""
echo "🔐 SSL Certificate will be automatically issued by Let's Encrypt"
echo "📊 Monitor certificate status: kubectl describe certificate api-example-com-tls"
