import {call, put, fork, take} from 'redux-saga/effects';
import {eventChannel} from 'redux-saga';
import {actions} from './actions';

function* init() {
  yield fork(connectWS);
}

function* connectWS() {
  const url = `${window.location.protocol=='https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws`;
  const channel = yield call(wsChannel, url);
  while (true) {
    const data = yield take(channel);
    yield put(actions.addEventMessage(data));
  }
}

function wsChannel(url) {
  const channel = eventChannel((emit)=>{
    const socket = new WebSocket(url);
    function onMessage(e) {
      emit(JSON.parse(e.data));
    }
    socket.addEventListener('message', onMessage);
    socket.addEventListener('error', ()=>{
      socket.close();
    });
    socket.addEventListener('close', ()=>{
    });
    return ()=>{
      socket.removeEventListener('message', onMessage);
    };
  });
  return channel;
}

function* rootSaga() {
  yield call(init);
}

export default rootSaga;
