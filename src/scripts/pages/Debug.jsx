import React from 'react';
import container from '../container';

class Debug extends React.Component {
  render() {
    return (
      <div>
        <p>{JSON.stringify(this.props.topology)}</p>
        {Object.keys(this.props.networkState).map((id)=>{
          if (this.props.networkState[id].pollSuccess) {
            return (
              <div key={id}>
                <h1>{id}: {this.props.networkState[id].hostname}</h1>
                <table>
                  <tbody>
                    <tr>
                      <th>Name</th>
                      <th>Type</th>
                      <th>OperStatus</th>
                      <th>hcInOctets</th>
                      <th>hcOutOctets</th>
                      <th>inOctets</th>
                      <th>outOctets</th>
                      <th>inDiscards</th>
                      <th>outDiscards</th>
                      <th>inErrors</th>
                      <th>outErrors</th>
                    </tr>
                    {Object.keys(this.props.networkState[id].interfaces).map((ifIndex)=>(
                      <tr key={ifIndex}>
                        <th>{this.props.networkState[id].interfaces[ifIndex].name}</th>
                        <td>{this.props.networkState[id].interfaces[ifIndex].type}</td>
                        <td>{this.props.networkState[id].interfaces[ifIndex].operStatus}</td>
                        <td>{this.props.networkState[id].interfaces[ifIndex].traffic.hcInOctets}</td>
                        <td>{this.props.networkState[id].interfaces[ifIndex].traffic.hcOutOctets}</td>
                        <td>{this.props.networkState[id].interfaces[ifIndex].traffic.inOctets}</td>
                        <td>{this.props.networkState[id].interfaces[ifIndex].traffic.outOctets}</td>
                        <td>{this.props.networkState[id].interfaces[ifIndex].traffic.inDiscards}</td>
                        <td>{this.props.networkState[id].interfaces[ifIndex].traffic.outDiscards}</td>
                        <td>{this.props.networkState[id].interfaces[ifIndex].traffic.inErrors}</td>
                        <td>{this.props.networkState[id].interfaces[ifIndex].traffic.outErrors}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            );
          } else {
            return (
              <div key={id}>
                <h1>{id}</h1>
              </div>
            );
          }
        })}
      </div>
    );
  }
}

export default container(Debug);
