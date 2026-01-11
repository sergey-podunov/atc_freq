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

// Tab switching function
window.switchTab = function (tabName: string) {
    // Hide all tab contents
    document.querySelectorAll('.tab-content').forEach(tab => {
        (tab as HTMLElement).style.display = 'none';
    });

    // Remove active class from all tabs
    document.querySelectorAll('.tab').forEach(tab => {
        tab.classList.remove('active');
    });

    // Show selected tab content
    const selectedTab = document.getElementById(tabName);
    if (selectedTab) {
        selectedTab.style.display = 'block';
    }

    // Add active class to clicked tab
    const clickedTab = document.querySelector(`[onclick="switchTab('${tabName}')"]`);
    if (clickedTab) {
        clickedTab.classList.add('active');
    }
};

document.querySelector('#app')!.innerHTML = `
    <div class="tabs">
        <button class="tab active" onclick="switchTab('frequencies')">Frequencies</button>
        <button class="tab" onclick="switchTab('weather')">Weather</button>
    </div>

    <div id="frequencies" class="tab-content">
        <div class="container">
          <div class="input-box" id="input">
            <input class="input" id="icao" type="text" autocomplete="off" placeholder="Enter ICAO (e.g. EDDB)" />
            <button class="btn" onclick="getFreq()">Get freq</button>
          </div>
          <div class="result" id="result">Results will appear here</div>
        </div>
    </div>

    <div id="weather" class="tab-content" style="display: none;">
        <div class="container">
          <div class="input-box">
            <input class="input" id="waypoint" type="text" autocomplete="off" placeholder="Enter waypoint" />
            <button class="btn" onclick="getWeather()">Get</button>
          </div>
          <div class="result" id="weather-result">Weather will appear here</div>
        </div>
    </div>
`;

// Get weather function (does nothing for now)
window.getWeather = function () {
    console.log('Get weather clicked');
};

let icaoElement = (document.getElementById("icao") as HTMLInputElement);
icaoElement.focus();
icaoElement.addEventListener("input", () => {
    icaoElement.value = icaoElement.value.toUpperCase();
});
icaoElement.addEventListener("keydown", (e) => {
    if (e.key === "Enter") {
        window.getFreq();
    }
});
let resultElement = document.getElementById("result");

declare global {
    interface Window {
        getFreq: () => void;
        getWeather: () => void;
        switchTab: (tabName: string) => void;
    }
}
