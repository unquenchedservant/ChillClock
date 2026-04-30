pragma ComponentBehavior: Bound

import QtQuick
import QtQuick.Layouts
import org.kde.plasma.plasmoid
import org.kde.plasma.components 3.0 as PlasmaComponents
import org.kde.plasma.plasma5support as Plasma5Support

PlasmoidItem {
  id: root

  property string timerText: "..."
  property color timerColor: "#dddddd"

  readonly property var colorMap: ({
      "green":  "#4caf50",
      "yellow": "#f9a825",
      "red":    "#ef5350",
      "white":  "#ffffff"
  })

  preferredRepresentation: fullRepresentation

  fullRepresentation: Item {
    Layout.minimumWidth: timerLabel.implicitWidth + 8
    Layout.fillHeight: true

    MouseArea {
      anchors.fill: parent
      acceptedButtons: Qt.LeftButton | Qt.RightButton
      onClicked: function(mouse) {
        if (mouse.button === Qt.RightButton)
          executable.connectSource("touch $HOME/dhv_timer_click2")
        else
          executable.connectSource("touch $HOME/dhv_timer_click1")
      }
    }

    PlasmaComponents.Label {
      id: timerLabel
      anchors.centerIn: parent
      text: root.timerText
      color: root.timerColor
      font.bold: true
    }
  }

  Plasma5Support.DataSource {
    id: executable
    engine: "executable"
    connectedSources: []
    onNewData: function(sourceName, data) {
      disconnectSource(sourceName)
      if (!sourceName.startsWith("cat ")) return
      try {
        var d = JSON.parse(data["stdout"])
        root.timerText = d.text ?? "0:00"
        root.timerColor = root.colorMap[d["class"]] ?? "#dddddd"
      } catch(e) {
        root.timerText = "ERR"
        root.timerColor = "#ef5350"
      }
    }
  }

  Timer {
    interval: 1000
    running: true
    repeat: true
    triggeredOnStart: true
    onTriggered: executable.connectSource("cat $HOME/dhv_timer.txt")
  }
}
