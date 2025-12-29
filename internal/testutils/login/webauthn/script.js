document.getElementById('registerButton').addEventListener('click', register);
document.getElementById('loginButton').addEventListener('click', login);


function showMessage(message, isError = false) {
    const messageElement = document.getElementById('message');
    messageElement.textContent = message;
    messageElement.style.color = isError ? 'red' : 'green';
}

async function register() {
    // Retrieve the username from the input field
    const username = document.getElementById('username').value;
    const fullName = document.getElementById('name').value;

    try {
        // Get registration options from your server. Here, we also receive the challenge.
        const response = await fetch('http://localhost:17608/v1/registration/options', {
            method: 'POST', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email: username, name: fullName }),
            credentials: "include",
        });

        // Check if the registration options are ok.
        if (!response.ok) {
            const msg = await response.json();
            throw new Error('User already exists or failed to get registration options from server: ' + msg);
        }

        // Convert the registration options to JSON.
        const options = await response.json();

        // This triggers the browser to display the passkey / WebAuthn modal (e.g. Face ID, Touch ID, Windows Hello).
        // A new attestation is created. This also means a new public-private-key pair is created.
        let attestationResponse;
        try {
            attestationResponse = await SimpleWebAuthnBrowser.startRegistration(options.publicKey);
        } catch (error) {
            // Some basic error handling
            if (error.name === 'InvalidStateError') {
                console.log('Error: Authenticator was probably already registered by user')
            } else {
                console.error(error);
            }

            throw error;
        }

        // Send attestationResponse back to server for verification and storage.
        const verificationResponse = await fetch('http://localhost:17608/v1/registration/verification', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(attestationResponse),
            credentials: "include",
        });

        if (!verificationResponse.ok) {
            const msg = await verificationResponse.json();
            throw new Error('Failed to register: ' + msg);
        }


        const msg = await verificationResponse.json();
        showMessage("Success!", false)
    } catch
    (error) {
        showMessage('Error: ' + error.message, true);
    }
}

async function login() {
    // Retrieve the username from the input field
    const username = document.getElementById('username').value;

    try {
        // Get login options from your server. Here, we also receive the challenge.
        const response = await fetch('http://localhost:17608/v1/authentication/options', {
            method: 'POST', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email: username }),
            credentials: "include",
        });

        // Check if the login options are ok.
        if (!response.ok) {
            const msg = await response.json();
            throw new Error('Failed to get login options from server: ' + msg);
        }

        // Convert the login options to JSON.
        const options = await response.json();

        // This triggers the browser to display the passkey / WebAuthn modal (e.g. Face ID, Touch ID, Windows Hello).
        // A new assertionResponse is created. This also means that the challenge has been signed.
        let assertionResponse;
        try {
            assertionResponse = await SimpleWebAuthnBrowser.startAuthentication(options.publicKey);
        } catch (error) {
            console.log(error)
        }

        // Send assertionResponse back to server for verification.
        const verificationResponse = await fetch('http://localhost:17608/v1/authentication/verification', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(assertionResponse),
            credentials: "include",
        });


        if (!verificationResponse.ok) {
            const msg = await verificationResponse.json();
            throw new Error('Failed to login: ' + msg);
        }

        const msg = await verificationResponse.json();
        showMessage("Success!", false)

    } catch (error) {
        showMessage('Error: ' + error.message, true);
    }
}
