var apiBasePath = API_BASE_PATH
window.onload = function () {
    google.accounts.id.initialize({
        client_id: "75641163784-vjcmhnrc3so137989q28a924rb6d8676.apps.googleusercontent.com",
        callback: handleCredentialResponse,
        auto_select: true
    });
    google.accounts.id.renderButton(
        document.getElementById("signin-btn"),
        { theme: "outline", size: "large" }  // customization attributes
    );
    google.accounts.id.prompt(); // also display the One Tap dialog
}

function handleCredentialResponse(response) {
    console.log(response);
    window.authToken = response.credential
    getUserInfo()
}

function getUserInfo() {
    var opts = {
        method: "GET",
        headers: {
            "Authorization": "Bearer " + window.authToken
        },
    };
    fetch(apiBasePath + '/api/v1/users/info', opts).then(function (response) {
        return response.json()
    }).then(body => {
        document.getElementById("user-email").innerHTML = body.email;
        document.getElementById("user-roles").innerHTML = JSON.stringify(body.roles);
    }).catch(error => console.error(error));
}

function callAPI(path, method) {
    var opts = {
        method: method,
        headers: {
            "Authorization": "Bearer " + window.authToken
        },
    };
    fetch(apiBasePath + '/api/v1/' + path, opts).then(function (response) {
        let result = {
            status: response.status,
            method: method,
            server: response.headers.get('X-Server')
        }
        if (response.ok) {
            response.json().then(body => {
                result.data = body;
                document.getElementById("api-result").innerHTML = JSON.stringify(result, null, "  ");
            }).catch(error => {
                result.data = error;
                document.getElementById("api-result").innerHTML = JSON.stringify(result, null, "  ");
            })
        } else {
            document.getElementById("api-result").innerHTML = JSON.stringify(result, null, "  ")
        }
    }).catch(error => console.error(error));
}
