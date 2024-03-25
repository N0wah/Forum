import { Application } from '@splinetool/runtime';

const canvas = document.getElementById('canvas3d');
const app = new Application(canvas);
app.load('https://prod.spline.design/33JVBKmHMCb0cZ98/scene.splinecode');

document.addEventListener("DOMContentLoaded", function() {
    var divClicable = document.getElementById("creatTopic");

    // Ajoute un écouteur d'événements pour le clic sur l'élément div
    divClicable.addEventListener("click", function() {
        // Redirige l'utilisateur vers la nouvelle page
        window.location.href = "creaTopic.html";
    });
});