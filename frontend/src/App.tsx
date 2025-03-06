import { Tab } from "./components";
import { BoilerData } from "./types";
import { useEffect, useState } from "react";
import {
  Connect,
  LoadConfig,
  SaveConfig,
  PLC_DATA,
} from "../wailsjs/go/main/App";
import { socket } from "./socket";

export default function App() {
  const [isModbusConnected, setIsModbusConnected] = useState<boolean>(false);
  const [configs, setConfigs] = useState({
    comPort: "",
    plantId: "",
  });

  //setBoilerData
  const [boilerData, setBoilerData] = useState<BoilerData[]>([
    {
      id: 0,
      reactorTemp: 0,
      separatorTemp: 0,
      furnaceTemp: 0,
      condenserTemp: 0,
      atmTemp: 0,
      reactorPressure: 0,
      gasTankPressure: 0,
      processStartTime: 0,
      timeOfReaction: 0,
      processEndTime: 0,
      coolingEndTime: 0,
      nitrogenPurging: 0,
      carbonDoorStatus: 0,
      coCh4Leakage: 0,
      jaaliBlockage: 0,
      machineMaintenance: 0,
      autoShutDown: 0,
    },
  ]);

  // socket data
  const [isSocketConnected, setIsSocketConnected] = useState<boolean>(
    socket.connected
  );

  // handle save
  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const plantId = (form.elements[0] as HTMLInputElement).value;
    const comPort = (form.elements[1] as HTMLInputElement).value;
    await SaveConfig({ comPort, plantId });
  };

  // load & save configs
  useEffect(() => {
    LoadConfig().then((data) => {
      setConfigs(data);
    });
  }, [setConfigs]);

  // socket config
  useEffect(() => {
    const handleConnect = () => {
      setIsSocketConnected(true);
    };

    const handleDisconnect = () => {
      setIsSocketConnected(false);
    };

    socket.on("connect", handleConnect);

    socket.on("connect_error", (error: Error) => {
      console.error("Socket.IO connect error:", error.message);
    });

    socket.on("disconnect", handleDisconnect);

    if (!socket.connected) {
      socket.connect();
    }

    return () => {
      socket.off("connect", handleConnect);
      socket.off("connect_error");

      socket.off("disconnect", handleDisconnect);
      socket.disconnect();
    };
  }, []);

  // load data
  useEffect(() => {
    let interval: number | undefined;

    const fetchAndSendData = async () => {
      if (!isModbusConnected) {
        console.log("Modbus not connected, skipping fetch");
        setBoilerData([]);
        return;
      }

      try {
        const data = await PLC_DATA();
        setBoilerData([]);

        if (isSocketConnected) {
          socket.emit("realtime", data);
        } else {
          console.log("Socket not connected, skipping send");
        }
      } catch (error) {
        console.error("Error fetching data from Go:", error);
        setIsModbusConnected(false);
      }
    };

    if (isModbusConnected) {
      fetchAndSendData();
      interval = setInterval(fetchAndSendData, 2000) as unknown as number;
    }

    return () => {
      if (interval !== undefined) {
        clearInterval(interval);
      }
    };
  }, [isModbusConnected, isSocketConnected]);

  return (
    <section className="max-w-7xl mx-auto p-5">
      {/* first part */}
      <form
        onSubmit={handleSubmit}
        className="flex items-center gap-10 justify-center"
      >
        <input
          defaultValue={configs.plantId}
          type="text"
          className="outline-none p-2 bg-white rounded-md w-md"
          placeholder="Please Enter Plant ID"
        />
        <input
          defaultValue={configs.comPort}
          type="text"
          className="outline-none p-2 bg-white rounded-md w-md"
          placeholder="Please Enter COM PORT i.e COM9"
        />
        <button
          type="submit"
          className="bg-gray-700 text-white px-4 py-2 rounded-md font-medium uppercase disabled:bg-gray-400"
        >
          Save
        </button>
      </form>

      {/* second part */}
      <div className="flex items-center justify-center gap-10 mt-5">
        <h1
          className={`text-center text-xl font-bold ${
            isModbusConnected ? "text-green-600" : "text-red-700"
          }`}
        >
          {isModbusConnected ? "Connected" : "Disconnected"}
        </h1>

        <button
          disabled={isModbusConnected}
          className="bg-gray-700 text-white px-4 py-2 rounded-md font-medium uppercase disabled:bg-gray-400"
          onClick={async () => {
            const res = await Connect(configs.plantId, configs.comPort);
            setIsModbusConnected(res);
          }}
        >
          Connect
        </button>
      </div>

      {/* third pard */}
      <div className="flex items-center justify-around gap-8 flex-wrap mt-8">
        {boilerData.map((data, index) => (
          <div key={index} className="border-2 border-black rounded p-3 w-sm">
            <div className="flex items-center gap-3 justify-center">
              <h1 className="text-3xl font-semibold uppercase">
                Boiler {data.id}
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
            {/*  */}
            <div className="flex flex-col mt-3 p-2 gap-3">
              <h3 className="uppercase text-center text-2xl font-semibold">
                Pressure Reading
              </h3>
              <Tab label="Reactor Pressure">{data.reactorPressure}</Tab>
              <Tab label="Gas Tank Pressure">{data.gasTankPressure}</Tab>
            </div>
            {/*  */}
            <div className="flex flex-col mt-3 p-2 gap-3">
              <h3 className="uppercase text-center text-2xl font-semibold">
                Operational Outputs
              </h3>
              <Tab label="Process Start Time">{data.processStartTime}</Tab>
              <Tab label="Time of Reaction">{data.timeOfReaction}</Tab>
              <Tab label="Process End Time">{data.processEndTime}</Tab>
              <Tab label="Cooling End Time">{data.coolingEndTime}</Tab>
            </div>
            {/*  */}
            <div className="flex flex-col mt-3 p-2 gap-3">
              <h3 className="uppercase text-center text-2xl font-semibold">
                Operational Indication
              </h3>
              <Tab label="Nitrogen Purging">
                <span
                  className={`rounded py-0.5 px-3 text-white ${
                    data.nitrogenPurging === 0 ? "bg-red-900" : "bg-green-900"
                  }`}
                >
                  {data.nitrogenPurging === 0 ? "red" : "green"}
                </span>
              </Tab>
              <Tab label="Carbon Door Status">
                <span
                  className={`rounded py-0.5 px-3 text-white ${
                    data.carbonDoorStatus === 0 ? "bg-red-900" : "bg-green-900"
                  }`}
                >
                  {data.carbonDoorStatus === 0 ? "red" : "green"}
                </span>
              </Tab>
            </div>
            {/*  */}
            <div className="flex flex-col mt-3 p-2 gap-3">
              <h3 className="uppercase text-center text-2xl font-semibold">
                Safety Indication
              </h3>
              <Tab label="Co-ch4 Gas Leakage">
                <span
                  className={`rounded py-0.5 px-3 text-white ${
                    data.coCh4Leakage !== 0 ? "bg-red-900" : "bg-green-900"
                  }`}
                >
                  {data.coCh4Leakage !== 0 ? "red" : "green"}
                </span>
              </Tab>
              <Tab label="Jaali Blockage">
                <span
                  className={`rounded py-0.5 px-3 text-white ${
                    data.jaaliBlockage !== 0 ? "bg-red-900" : "bg-green-900"
                  }`}
                >
                  {data.jaaliBlockage !== 0 ? "red" : "green"}
                </span>
              </Tab>
              <Tab label="Machine Maintenance">
                <span
                  className={`rounded py-0.5 px-3 text-white ${
                    data.machineMaintenance !== 0
                      ? "bg-red-900"
                      : "bg-green-900"
                  }`}
                >
                  {data.machineMaintenance !== 0 ? "red" : "green"}
                </span>
              </Tab>
              <Tab label="Auto Shut Down">
                <span
                  className={`rounded py-0.5 px-3 text-white ${
                    data.autoShutDown !== 0 ? "bg-red-900" : "bg-green-900"
                  }`}
                >
                  {data.autoShutDown !== 0 ? "red" : "green"}
                </span>
              </Tab>
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}
