import './style.css';
import './app.css';

import {GetFrequencies} from '../wailsjs/go/app/App';
import {sim} from '../wailsjs/go/models';

// Setup the getFreq function
window.getFreq = function () {
    // Get ICAO
    let icao = icaoElement!.value;

    // Check if the input is empty
    if (icao === "") return;

    resultElement!.innerText = "Fetching frequencies for " + icao + "...";

    // Call App.GetFrequencies(icao)
    try {
        GetFrequencies(icao)
            .then((result: sim.AirportFrequency[]) => {
                if (result && result.length > 0) {
                    let html = '<table style="width:100%; text-align: left; border-collapse: collapse;">';
                    html += '<tr><th>Type</th><th>MHz</th><th>Name</th></tr>';
                    result.forEach((f: sim.AirportFrequency) => {
                        html += `<tr>
                            <td>${f.Type}</td>
                            <td>${f.MHz.toFixed(3)}</td>
                            <td>${f.Name}</td>
                        </tr>`;
                    });
                    html += '</table>';
                    resultElement!.innerHTML = html;
                } else {
                    resultElement!.innerText = "No frequencies found or error occurred.";
                }
            })
            .catch((err: any) => {
                console.error(err);
                resultElement!.innerText = "Error: " + err;
            });
    } catch (err: any) {
        console.error(err);
        resultElement!.innerText = "Error: " + err;
    }
};

document.querySelector('#app')!.innerHTML = `
    <div class="container">
      <div class="input-box" id="input">
        <input class="input" id="icao" type="text" autocomplete="off" placeholder="Enter ICAO (e.g. EDDB)" />
        <button class="btn" onclick="getFreq()">Get freq</button>
      </div>
      <div class="result" id="result">Results will appear here</div>
    </div>
`;

let icaoElement = (document.getElementById("icao") as HTMLInputElement);
icaoElement.focus();
let resultElement = document.getElementById("result");

declare global {
    interface Window {
        getFreq: () => void;
    }
}
