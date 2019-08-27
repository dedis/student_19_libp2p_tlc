from matplotlib import pyplot as plt

if __name__ == '__main__':
    Filename = "LeaveRejoin_main"
    with open("../logs/" + Filename + ".log") as f:
        line = f.readline()
        data = dict()
        while line:
            parts = line.split()
            node, timeStep = parts[2].split(",")
            _, millisec = parts[1].split(".")
            if node not in data.keys():
                times = list()
            else:
                times = data[node]
            times.append((int(timeStep), float(parts[0] + "." + millisec)))
            data[node] = times
            line = f.readline()

    data["0"], data["4"] = data["4"], data["0"]
    data["1"], data["3"] = data["3"], data["1"]

    base = min([a[0][1] for a in list(data.values())])
    for j,node in enumerate(data.keys()):
        print(node)
        plt.step([i[1] - base for i in data[node]], [i[0] + 0.03*j for i in data[node]], where='post')

    plt.xlabel("time in seconds")
    plt.ylabel("time steps")
    # plt.savefig("../graphs/" + Filename + ".png")
    plt.show()
