Oidc.Log.logger = console;
Oidc.Log.level = Oidc.Log.INFO;

const settings = {
  authority: 'https://connect-staging.fi-group.com/identity',
  client_id: 'fileblobs',
  redirect_uri: window.location.origin + '/login',
  response_type: 'id_token token',
  scope: 'openid profile email',
  filterProtocolClaims: true,
  loadUserInfo: true
};

const mgr = new Oidc.UserManager(settings);

mgr.events.addUserLoaded(function (login) {
  console.log("Usuário carregado:", login);
});

function handleAuthError(err) {
  console.error("Erro de autenticação:", err);
  
  // Verifica se é um erro de acesso negado (403 Forbidden)
  if (err.error === "access_denied" || err.status === 403) {
    console.log("Acesso negado detectado - redirecionando para página de acesso negado");
    
    // Cria um cookie para evitar loop de redirecionamento
    document.cookie = "access_denied=true; path=/";
    
    // Redireciona para a página de acesso negado
    let redirectUrl = "/access-denied";
    
    // Adiciona a mensagem de erro à URL se disponível
    if (err.error_description) {
      redirectUrl += "?message=" + encodeURIComponent(err.error_description);
    }
    
    window.location.href = redirectUrl;
    return;
  }
  
  // Para outros erros, exibe a mensagem no elemento loginResult
  const loginResult = document.getElementById("loginResult");
  if (loginResult) {
    loginResult.innerHTML = "Erro de autenticação: " + (err.error_description || err.message || "Erro desconhecido");
  }
}

function processLoginResponse() {
  console.log("Processing login response...");
  console.log("URL search:", window.location.search);  mgr.signinRedirectCallback().then(login => {
      console.log("Login successful, token received:", login.access_token.substring(0, 10) + "...");
      fetch("/auth/store-token", {
        method: "POST",
        headers: { 
          "Content-Type": "application/json",
          // Adicionar cabeçalho para indicar que é uma solicitação JavaScript
          "X-Requested-With": "XMLHttpRequest"
        },
        credentials: "include",
        body: JSON.stringify({ token: login.access_token })
      })
      .then(response => {
        console.log("Token store response:", response.status);
        
        // Verificar explicitamente por status 403 (Forbidden)
        if (response.status === 403) {
          console.log("Acesso negado - redirecionando para página de acesso negado");
          document.cookie = "access_denied=true; path=/";
          window.location.href = "/access-denied";
          throw new Error("Redirecionando para página de acesso negado");
        }
        
        if (!response.ok) {
          return response.json().then(data => {
            throw data; // Passa o objeto de erro completo
          });
        }
        return response.json();
      })
      .then(data => {
        console.log("Token store data:", data);
        window.location.href = "/storage-accounts";
      })
      .catch(err => {
        console.error("Erro no processamento do token:", err);
        
        // Se o erro for o redirecionamento que já fizemos, não faz nada
        if (err.message === "Redirecionando para página de acesso negado") {
          return;
        }
        
        if (typeof handleAuthError === 'function') {
          handleAuthError(err);
        } else {
          document.getElementById("loginResult").innerHTML = "Erro de autenticação: " + err.message;
        }
      });
    })
    .catch(err => {
      console.error("Erro no login:", err);
      if (typeof handleAuthError === 'function') {
        handleAuthError(err);
      } else {
        document.getElementById("loginResult").innerHTML = "Erro de autenticação: " + err.message;
      }
    });
}

if (window.location.search || window.location.hash) {
  console.log("Found URL parameters, processing login response...");
  document.getElementById("loginResult").innerHTML = "Processando resposta de login...";
  processLoginResponse();
} else {
  console.log("No URL parameters, checking user state...");
  document.getElementById("loginResult").innerHTML = "Verificando estado do usuário...";
  mgr.getUser().then(login => {
    if (!login || login.expired) {
      console.log("Sem token válido. Redirecionando para login...");
      document.getElementById("loginResult").innerHTML = "Redirecionando para autenticação...";
      mgr.signinRedirect().catch(err => {
        console.error("Erro ao redirecionar para login:", err);
        document.getElementById("loginResult").innerHTML = "Erro ao redirecionar para autenticação: " + err.message;
      });
    } else {      console.log("Usuário já autenticado, enviando token...");
      document.getElementById("loginResult").innerHTML = "Usuário autenticado, processando...";
      
      fetch("/auth/store-token", {
        method: "POST",
        headers: { 
          "Content-Type": "application/json",
          // Adicionar cabeçalho para indicar que é uma solicitação JavaScript
          "X-Requested-With": "XMLHttpRequest"
        },
        credentials: "include",
        body: JSON.stringify({ token: login.access_token })
      })
      .then(response => {
        console.log("Token store response:", response.status);
        
        // Verificar explicitamente por status 403 (Forbidden)
        if (response.status === 403) {
          console.log("Acesso negado - redirecionando para página de acesso negado");
          document.cookie = "access_denied=true; path=/";
          window.location.href = "/access-denied";
          throw new Error("Redirecionando para página de acesso negado");
        }
        
        if (!response.ok) {
          return response.json().then(data => {
            throw data; // Passa o objeto de erro completo
          });
        }
        return response.json();
      })
      .then(data => {
        console.log("Token store data:", data);
        window.location.href = "/storage-accounts";
      })
      .catch(err => {
        console.error("Erro ao armazenar token:", err);
        
        // Se o erro for o redirecionamento que já fizemos, não faz nada
        if (err.message === "Redirecionando para página de acesso negado") {
          return;
        }
        
        if (typeof handleAuthError === 'function') {
          handleAuthError(err);
        } else {
          document.getElementById("loginResult").innerHTML = "Erro de autenticação: " + (err.error_description || err.message || "Erro desconhecido");
        }
      });
    }  }).catch(err => {
    console.error("Erro ao verificar usuário:", err);
    document.getElementById("loginResult").innerHTML = "Erro ao verificar usuário: " + err.message;
  });
}

