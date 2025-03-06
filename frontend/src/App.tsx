import { useEffect, useState } from "react";
import { Connect, LoadConfig, SaveConfig } from "../wailsjs/go/main/App";

export default function App() {
  const [isModbusConnected, setIsModbusConnected] = useState<boolean>(false);
  const [configs, setConfigs] = useState({
    comPort: "",
    plantId: "",
  });

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const plantId = (form.elements[0] as HTMLInputElement).value;
    const comPort = (form.elements[1] as HTMLInputElement).value;
    await SaveConfig({ comPort, plantId });
  };

  useEffect(() => {
    LoadConfig().then((data) => {
      setConfigs(data);
    });
  }, [setConfigs]);

  return (
    <section className="max-w-7xl mx-auto p-5">
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
    </section>
  );
}
