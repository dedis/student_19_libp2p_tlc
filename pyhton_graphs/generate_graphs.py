from matplotlib import pyplot as plt

if __name__ == '__main__':
    Filename = "NoFailure_NoDelay"
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

    for node in data.keys():
        print(node)
        plt.step([i[1] for i in data[node]], [i[0] for i in data[node]], where='post')

    plt.savefig("../graphs/" + Filename + ".png")
    plt.show()
