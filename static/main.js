const id = Math.random().toString(36).substring(2, 15)
const socket = new WebSocket(`ws://${location.host}/ws?id=${id}&token=1234`);
const messages = document.getElementById('main')
const textInput = document.getElementById('text')
const send = document.getElementById('send')

send.addEventListener('click', ()=>{
    const text = textInput.value
    if(text){
        socket.send(text) 
        textInput.value = ''
    }
})
console.log("Attempting Connection...");

socket.onopen = () => {
    console.log("Successfully Connected");
};

socket.onclose = event => {
    console.log("Socket Closed Connection: ", event);
};

socket.onerror = error => {
    console.log("Socket Error: ", error);
};

socket.onmessage = event =>{
    console.log("Socket on message ", event);
    const newMessage = document.createElement('div')
    newMessage.textContent = event.data
    messages.appendChild(newMessage)
}