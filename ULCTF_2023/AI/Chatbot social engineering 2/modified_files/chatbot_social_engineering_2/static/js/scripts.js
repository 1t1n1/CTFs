function sound(src) {
  this.sound = document.createElement("audio");
  this.sound.src = src;
  this.sound.setAttribute("preload", "auto");
  this.sound.setAttribute("controls", "none");
  this.sound.style.display = "none";
  document.body.appendChild(this.sound);
  this.play = function(){
    this.sound.play();
  }
  this.stop = function(){
    this.sound.pause();
  }
}

var speak1 = new sound('/static/audio/speak1.mp3');
var speak2 = new sound('/static/audio/speak2.mp3');
var speak3 = new sound('/static/audio/speak3.mp3');

function submitForm(userInput = document.getElementById("user_input").value) {
    // envoyer la requête au backend
    fetch('/submit', {
        method: 'POST',
        body: JSON.stringify({user_input: userInput}),
        headers: {
          'Content-Type': 'application/json'
        }
    })
    .then(response => response.text())
    .then(data => {
        // Afficher la réponse de l'API ici sous forme de bulle de conversation
        typeWriter(data, document.getElementById("chatbot_response"), 60, true);
    });

    // afficher le message dans la bulle
    var message_sent = document.getElementById('message_sent');
    typeWriter(userInput, message_sent);
    document.getElementById("user_input").value = "";
}

// add a typewriter effect to the message
function typeWriter(text, element, delay = 60, audio = false) {
    element.innerHTML = '';
    let i = 0;
    let timer = setInterval(function () {
        if (audio) {
            if (i % 4 == 0) {
                var audio_files = [speak1, speak2, speak3];
                audio_files[Math.floor(Math.random() * audio_files.length)].play();
            }
        }
        element.innerHTML += text[i];
        i++;
        if (i > text.length - 1) {
            clearInterval(timer);
        }
    }, delay);
}

// Initial message
submitForm('Bonjour !');