import { useEffect, useState } from "react";
import { Tab } from "./components";
import { socket } from "./socket";
import { BoilerData } from "./types";
import {
  GetBoilerData,
  Connect,
  IsModbusConnected,
} from "../wailsjs/go/main/App";

function App() {
  const [isSocketConnected, setIsSocketConnected] = useState<boolean>(
    socket.connected
  );
  const [isModbusConnected, setIsModbusConnected] = useState<boolean>(false);
  const [boilerData, setBoilerData] = useState<BoilerData[]>([]);
  const [connectionStatus, setConnectionStatus] = useState<string>("0"); // Default to "0"

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const plantId = (form.elements[0] as HTMLInputElement).value;
    const comPort = (form.elements[1] as HTMLInputElement).value;
    try {
      const result = await Connect(plantId, comPort);
      setConnectionStatus(result);
      const connected = await IsModbusConnected();
      setIsModbusConnected(connected);
      console.log("Connection result:", result, "Modbus connected:", connected);
    } catch (error) {
      setConnectionStatus("Error: " + (error as Error).message);
      setIsModbusConnected(false);
      console.error("Connection error:", error);
    }
  };

  useEffect(() => {
    console.log("Socket initial state:", socket.connected);

    const handleConnect = () => {
      setIsSocketConnected(true);
      console.log("Socket.IO connected");
    };

    const handleDisconnect = () => {
      setIsSocketConnected(false);
      console.log("Socket.IO disconnected");
    };

    socket.on("connect", handleConnect);
    socket.on("connect_error", (error: Error) => {
      console.error("Socket.IO connect error:", error.message);
    });
    socket.on("disconnect", handleDisconnect);
    socket.on("realtime", (data) => {
      console.log("Received from server:", data);
    });

    if (!socket.connected) {
      console.log("Attempting Socket.IO connection");
      socket.connect();
    }

    return () => {
      socket.off("connect", handleConnect);
      socket.off("disconnect", handleDisconnect);
      socket.off("connect_error");
      socket.off("realtime");
      socket.disconnect();
    };
  }, []);

  useEffect(() => {
    let interval: NodeJS.Timeout;

    const fetchAndSendData = async () => {
      if (!isModbusConnected) {
        console.log("Modbus not connected, skipping fetch");
        setBoilerData([]);
        setConnectionStatus("0"); // Reset to "0" when not connected
        return;
      }

      try {
        const data = await GetBoilerData();
        setBoilerData(data);
        console.log("Fetched data from Go:", data);

        if (isSocketConnected) {
          socket.emit("realtime", data);
          console.log("Sent data to Socket.IO server:", data);
        } else {
          console.log("Socket not connected, skipping send");
        }
      } catch (error) {
        console.error("Error fetching data from Go:", error);
        setIsModbusConnected(false);
        setConnectionStatus("Fetch error: " + (error as Error).message);
      }
    };

    if (isModbusConnected) {
      fetchAndSendData(); // Initial fetch
      interval = setInterval(fetchAndSendData, 2000);
    }

    return () => {
      if (interval) clearInterval(interval);
    };
  }, [isModbusConnected, isSocketConnected]);

  return (
    <section className="max-w-7xl mx-auto p-5">
      <form
        onSubmit={handleSubmit}
        className="flex items-center gap-10 justify-center"
      >
        <input
          type="text"
          className="outline-none p-2 bg-white rounded-md w-md"
          placeholder="Please Enter Plant ID"
        />
        <input
          type="text"
          className="outline-none p-2 bg-white rounded-md w-md"
          placeholder="Please Enter COM PORT i.e COM9"
        />
        <button
          disabled={isModbusConnected} // Disable when connected
          type="submit"
          className="bg-gray-700 text-white px-4 py-2 rounded-md font-medium uppercase disabled:bg-gray-400"
        >
          Connect
        </button>
      </form>
      <p>{connectionStatus}</p> {/* Always show status, defaults to "0" */}
      <div className="flex items-center justify-around gap-8 flex-wrap mt-8">
        {boilerData.map((data, index) => (
          <div key={index} className="border-2 border-black rounded p-3 w-sm">
            <div className="flex items-center gap-3 justify-center">
              <h1 className="text-3xl font-semibold uppercase">
                Boiler {index}
              </h1>
              <p className="font-medium uppercase">Active</p>
            </div>
            <div className="flex flex-col mt-3 p-2 gap-3">
              <h3 className="uppercase text-center text-2xl font-semibold">
                Temperature Reading
              </h3>
              <Tab label="Reactor Temperature">{data.reactorTemp}</Tab>
              <Tab label="Separator Temperature">{data.separatorTemp}</Tab>
              <Tab label="Furnace Temperature">{data.furnaceTemp}</Tab>
              <Tab label="Condenser Temperature">{data.condenserTemp}</Tab>
              <Tab label="Atmosphere Temperature">{data.atmTemp}</Tab>
            </div>
            <div className="flex flex-col mt-3 p-2 gap-3">
              <h3 className="uppercase text-center text-2xl font-semibold">
                Pressure Reading
              </h3>
              <Tab label="Reactor Pressure">{data.reactorPressure}</Tab>
              <Tab label="Gas Tank Pressure">{data.gasTankPressure}</Tab>
            </div>
            <div className="flex flex-col mt-3 p-2 gap-3">
              <h3 className="uppercase text-center text-2xl font-semibold">
                Operational Outputs
              </h3>
              <Tab label="Process Start Time">{data.processStartTime}</Tab>
              <Tab label="Time of Reaction">{data.timeOfReaction}</Tab>
              <Tab label="Process End Time">{data.processEndTime}</Tab>
              <Tab label="Cooling End Time">{data.coolingEndTime}</Tab>
            </div>
            <div className="flex flex-col mt-3 p-2 gap-3">
              <h3 className="uppercase text-center text-2xl font-semibold">
                Operational Indication
              </h3>
              <Tab label="Nitrogen Purging">
                <span
                  className={`rounded py-0.5 px-3 text-white ${
                    data.nitrogenPurging === "red"
                      ? "bg-red-900"
                      : "bg-green-900"
                  }`}
                >
                  {data.nitrogenPurging}
                </span>
              </Tab>
              <Tab label="Carbon Door Status">
                <span
                  className={`rounded py-0.5 px-3 text-white ${
                    data.carbonDoorStatus === "red"
                      ? "bg-red-900"
                      : "bg-green-900"
                  }`}
                >
                  {data.carbonDoorStatus}
                </span>
              </Tab>
            </div>
            <div className="flex flex-col mt-3 p-2 gap-3">
              <h3 className="uppercase text-center text-2xl font-semibold">
                Safety Indication
              </h3>
              <Tab label="Co-ch4 Gas Leakage">
                <span
                  className={`rounded py-0.5 px-3 text-white ${
                    data.coCh4Leakage === "red" ? "bg-red-900" : "bg-green-900"
                  }`}
                >
                  {data.coCh4Leakage}
                </span>
              </Tab>
              <Tab label="Jaali Blockage">
                <span
                  className={`rounded py-0.5 px-3 text-white ${
                    data.jaaliBlockage === "red" ? "bg-red-900" : "bg-green-900"
                  }`}
                >
                  {data.jaaliBlockage}
                </span>
              </Tab>
              <Tab label="Machine Maintenance">
                <span
                  className={`rounded py-0.5 px-3 text-white ${
                    data.machineMaintenance === "red"
                      ? "bg-red-900"
                      : "bg-green-900"
                  }`}
                >
                  {data.machineMaintenance}
                </span>
              </Tab>
              <Tab label="Auto Shut Down">
                <span
                  className={`rounded py-0.5 px-3 text-white ${
                    data.autoShutDown === "red" ? "bg-red-900" : "bg-green-900"
                  }`}
                >
                  {data.autoShutDown}
                </span>
              </Tab>
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}

export default App;
