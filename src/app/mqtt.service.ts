import {EventEmitter, Inject, Injectable, InjectionToken, Provider} from '@angular/core';
import { MqttClient, IClientOptions, IClientPublishOptions, IClientSubscribeOptions, ISubscriptionGrant, IConnackPacket, IStream} from 'mqtt';
import mqtt from "mqtt";
import { Packet } from 'mqtt-packet';

import { BehaviorSubject, merge, Observable, Observer, Subject, Subscription, Unsubscribable, using } from 'rxjs';
import { filter, publish, publishReplay, refCount } from 'rxjs/operators';

import {IPacket} from "mqtt-packet";

export enum MqttConnectionState {
  CLOSED,
  CONNECTING,
  CONNECTED
}

export interface IMqttServiceOptions extends IClientOptions {
  /**
   * whether a new connection should be created
   * on creating an instance of the service
   */
  connectOnCreate?: boolean;
  /** the hostname of the mqtt broker */
  hostname?: string;
  /** the port to connect with websocket to the broker */
  port?: number;
  /** the path parameters to connect to e.g. `/mqtt` */
  path?: string;
  protocol?: 'wss' | 'ws';
  /** if the url is provided, hostname, port path and protocol are ignored */
  url?: string;
}

export interface IMqttMessage extends IPacket {
  /** the mqtt topic to which this message was published to */
  topic: string;
  /** the payload */
  payload: string | Buffer;
  /** the quality of service */
  qos: number;
  /** if this message is a retained message */
  retain: boolean;
  /** if this message is a duplicate */
  dup: boolean;
}

export const MQTT_SERVICE_OPTIONS: IMqttServiceOptions = {
  connectOnCreate: true,
  hostname: 'localhost',
  port: 1884,
  path: ''
};

export const MqttServiceConfig = new InjectionToken<IMqttServiceOptions>('NgxMqttServiceConfig');
export const MqttClientService = new InjectionToken<IMqttClient>('NgxMqttClientService');


export interface IPublishOptions extends IClientPublishOptions { }
export interface IOnConnectEvent extends IConnackPacket { }
export interface IOnErrorEvent extends Error {
  type?: string;
}
export interface IOnMessageEvent extends IMqttMessage { }
export interface IOnSubackEvent {
  granted: boolean;
  filter: string;
}

export interface IMqttClient extends MqttClient {
  stream: IStream;
}

export interface IOnPacketsendEvent extends IPacket { }
export interface IOnPacketreceiveEvent extends IPacket { }

// A javascript function that takes two objects and merges them recursively
function mergeDeep(target: any, ...sources: any[]): any {
  if (!sources.length) {
    return target;
  }
  const source = sources.shift();

  if (isObject(target) && isObject(source)) {
    for (const key in source) {
      if (isObject(source[key])) {
        if (!target[key]) {
          Object.assign(target, { [key]: {} });
        }
        mergeDeep(target[key], source[key]);
      } else {
        Object.assign(target, { [key]: source[key] });
      }
    }
  }

  return mergeDeep(target, ...sources);
}

function isObject(item: any): item is Object {
  return item && typeof item === 'object' && !Array.isArray(item);
}

/**
 * With an instance of MqttService, you can observe and subscribe to MQTT in multiple places, e.g. in different components,
 * to only subscribe to the broker once per MQTT filter.
 * It also handles proper unsubscription from the broker, if the last observable with a filter is closed.
 */
@Injectable({
  providedIn: 'root',
})
export class MqttService {
  private client!: MqttClient;

  /**
   * The constructor needs [connection options]{@link IMqttServiceOptions} regarding the broker and some
   * options to configure behavior of this service, like if the connection to the broker
   * should be established on creation of this service or not.
   */
  constructor(
    @Inject(MqttServiceConfig) private options: IMqttServiceOptions,
    @Inject(MqttClientService) client?: MqttClient
  ) {
    if (options.connectOnCreate !== false) {
      this.connect({}, client);
    }

    this.state.subscribe();
  }

  /**
   * gets the _clientId
   */
  public get clientId() {
    return this._clientId;
  }

  /** An EventEmitter to listen to connect messages */
  public get onConnect(): EventEmitter<IOnConnectEvent> {
    return this._onConnect;
  }

  /** An EventEmitter to listen to reconnect messages */
  public get onReconnect(): EventEmitter<void> {
    return this._onReconnect;
  }

  /** An EventEmitter to listen to close messages */
  public get onClose(): EventEmitter<void> {
    return this._onClose;
  }

  /** An EventEmitter to listen to offline events */
  public get onOffline(): EventEmitter<void> {
    return this._onOffline;
  }

  /** An EventEmitter to listen to error events */
  public get onError(): EventEmitter<IOnErrorEvent> {
    return this._onError;
  }

  /** An EventEmitter to listen to close messages */
  public get onEnd(): EventEmitter<void> {
    return this._onEnd;
  }

  /** An EventEmitter to listen to message events */
  public get onMessage(): EventEmitter<Packet> {
    return this._onMessage;
  }

  /** An EventEmitter to listen to packetsend messages */
  public get onPacketsend(): EventEmitter<IOnPacketsendEvent> {
    return this._onPacketsend;
  }

  /** An EventEmitter to listen to packetreceive messages */
  public get onPacketreceive(): EventEmitter<IOnPacketreceiveEvent> {
    return this._onPacketreceive;
  }

  /** An EventEmitter to listen to suback events */
  public get onSuback(): EventEmitter<IOnSubackEvent> {
    return this._onSuback;
  }

  /** An actual connection state **/
  public get connectionState(): MqttConnectionState {
    return this.state.value;
  }

  /** a map of all mqtt observables by filter */
  public observables: { [filterString: string]: Observable<IMqttMessage> } = {};
  /** the connection state */
  public state: BehaviorSubject<MqttConnectionState> = new BehaviorSubject<MqttConnectionState>(MqttConnectionState.CLOSED);
  /** an observable of the last mqtt message */
  public messages: Subject<IMqttMessage> = new Subject<IMqttMessage>();

  private _clientId = this._generateClientId();
  private _connectTimeout = 10000;
  private _reconnectPeriod = 10000;
  private _url!: string;

  private _onConnect: EventEmitter<IOnConnectEvent> = new EventEmitter<IOnConnectEvent>();
  private _onReconnect: EventEmitter<void> = new EventEmitter<void>();
  private _onClose: EventEmitter<void> = new EventEmitter<void>();
  private _onOffline: EventEmitter<void> = new EventEmitter<void>();
  private _onError: EventEmitter<IOnErrorEvent> = new EventEmitter<IOnErrorEvent>();
  private _onEnd: EventEmitter<void> = new EventEmitter<void>();
  private _onMessage: EventEmitter<Packet> = new EventEmitter<Packet>();
  private _onSuback: EventEmitter<IOnSubackEvent> = new EventEmitter<IOnSubackEvent>();
  private _onPacketsend: EventEmitter<IOnPacketsendEvent> = new EventEmitter<IOnPacketsendEvent>();
  private _onPacketreceive: EventEmitter<IOnPacketreceiveEvent> = new EventEmitter<IOnPacketreceiveEvent>();

  /**
   * This static method shall be used to determine whether a MQTT
   * topic matches a given filter. The matching rules are specified in the MQTT
   * standard documentation and in the library test suite.
   *
   * @param  {string}  filter A filter may contain wildcards like '#' and '+'.
   * @param  {string}  topic  A topic may not contain wildcards.
   * @return {boolean}        true on match and false otherwise.
   */
  public static filterMatchesTopic(filterString: string, topic: string): boolean {
    if (filterString[0] === '#' && topic[0] === '$') {
      return false;
    }
    // Preparation: split and reverse on '/'. The JavaScript split function is sane.
    const fs = (filterString || '').split('/').reverse();
    const ts = (topic || '').split('/').reverse();
    // This function is tail recursive and compares both arrays one element at a time.
    const match = (): boolean => {
      // Cutting of the last element of both the filter and the topic using pop().
      const f = fs.pop();
      const t = ts.pop();
      switch (f) {
        // In case the filter level is '#', this is a match no matter whether
        // the topic is undefined on this level or not ('#' matches parent element as well!).
        case '#':
          return true;
        // In case the filter level is '+', we shall dive into the recursion only if t is not undefined.
        case '+':
          return t ? match() : false;
        // In all other cases the filter level must match the topic level,
        // both must be defined and the filter tail must match the topic
        // tail (which is determined by the recursive call of match()).
        default:
          return f === t && (f === undefined ? true : match());
      }
    };
    return match();
  }

  /**
   * connect manually connects to the mqtt broker.
   */
  public connect(opts?: IMqttServiceOptions, client?: MqttClient) {
    const options = mergeDeep(this.options || {}, opts);
    const protocol = options.protocol || 'ws';
    const hostname = options.hostname || 'localhost';
    if (options.url) {
      this._url = options.url;
    } else {
      this._url = `${protocol}://${hostname}`;
      this._url += options.port ? `:${options.port}` : '';
      this._url += options.path ? `${options.path}` : '';
    }
    this.state.next(MqttConnectionState.CONNECTING);
    const mergedOptions = mergeDeep({
      clientId: this._clientId,
      reconnectPeriod: this._reconnectPeriod,
      connectTimeout: this._connectTimeout
    }, options);

    if (this.client) {
      this.client.end(true);
    }

    if (!client) {
      this.client = mqtt.connect(this._url, mergedOptions);
    } else {
      this.client = client;
    }
    this._clientId = mergedOptions.clientId;

    this.client.on('connect', this._handleOnConnect);
    this.client.on('reconnect', this._handleOnReconnect);
    this.client.on('close', this._handleOnClose);
    this.client.on('offline', this._handleOnOffline);
    this.client.on('error', this._handleOnError);
    (this.client as any).stream.on('error', this._handleOnError);
    this.client.on('end', this._handleOnEnd);
    this.client.on('message', this._handleOnMessage);
    this.client.on('packetsend', this._handleOnPacketsend);
    this.client.on('packetreceive', this._handleOnPacketreceive);
  }

  /**
   * disconnect disconnects from the mqtt client.
   * This method `should` be executed when leaving the application.
   */
  public disconnect(force = true) {
    if (!this.client) {
      throw new Error('mqtt client not connected');
    }
    this.client.end(force);
  }

  /**
   * With this method, you can observe messages for a mqtt topic.
   * The observable will only emit messages matching the filter.
   * The first one subscribing to the resulting observable executes a mqtt subscribe.
   * The last one unsubscribing this filter executes a mqtt unsubscribe.
   * Every new subscriber gets the latest message.
   */
  public observeRetained(filterString: string, opts: IClientSubscribeOptions = { qos: 1 }): Observable<IMqttMessage> {
    return this._generalObserve(filterString, () => publishReplay(1), opts);
  }

  /**
   * With this method, you can observe messages for a mqtt topic.
   * The observable will only emit messages matching the filter.
   * The first one subscribing to the resulting observable executes a mqtt subscribe.
   * The last one unsubscribing this filter executes a mqtt unsubscribe.
   */
  public observe(filterString: string, opts: IClientSubscribeOptions = { qos: 1 }): Observable<IMqttMessage> {
    return this._generalObserve(filterString, () => publish(), opts);
  }

  /**
   * With this method, you can observe messages for a mqtt topic.
   * The observable will only emit messages matching the filter.
   * The first one subscribing to the resulting observable executes a mqtt subscribe.
   * The last one unsubscribing this filter executes a mqtt unsubscribe.
   * Depending on the publish function, the messages will either be replayed after new
   * subscribers subscribe or the messages are just passed through
   */
  private _generalObserve(filterString: string, publishFn: Function, opts: IClientSubscribeOptions): Observable<IMqttMessage> {
    if (!this.client) {
      throw new Error('mqtt client not connected');
    }
    if (!this.observables[filterString]) {
      const rejected: Subject<IMqttMessage> = new Subject();
      this.observables[filterString] = using(
        // resourceFactory: Do the actual ref-counting MQTT subscription.
        // refcount is decreased on unsubscribe.
        () => {
          const subscription: Subscription = new Subscription();
          this.client.subscribe(filterString, opts, (err: any, granted: ISubscriptionGrant[] | undefined) => {
            if (granted) { // granted can be undefined when an error occurs when the client is disconnecting
              granted.forEach((granted_: ISubscriptionGrant) => {
                if (granted_.qos === 128) {
                  delete this.observables[granted_.topic];
                  this.client.unsubscribe(granted_.topic);
                  rejected.error(`subscription for '${granted_.topic}' rejected!`);
                }
                this._onSuback.emit({ filter: filterString, granted: granted_.qos !== 128 });
              });
            }
          });
          subscription.add(() => {
            delete this.observables[filterString];
            this.client.unsubscribe(filterString);
          });
          return subscription;
        },
        // observableFactory: Create the observable that is consumed from.
        // This part is not executed until the Observable returned by
        // `observe` gets actually subscribed.
        (subscription: Unsubscribable | void) => merge(rejected, this.messages))
        .pipe(
          filter((msg: IMqttMessage) => MqttService.filterMatchesTopic(filterString, msg.topic)),
          publishFn(),
          refCount()
        ) as Observable<IMqttMessage>;
    }
    return this.observables[filterString];
  }

  /**
   * This method returns an observable for a topic with optional options.
   * After subscribing, the actual mqtt publication will be executed and
   * the observable will emit an empty value and completes, if publishing was successful
   * or throws an error, if the publication fails.
   */
  public publish(topic: string, message: string | Buffer, options: IClientPublishOptions = {}): Observable<void> {
    if (!this.client) {
      throw new Error('mqtt client not connected');
    }
    return Observable.create((obs: Observer<void>) => {
      this.client.publish(topic, message, options, (error: Error | undefined) => {
        if (error) {
          obs.error(error);
        } else {
          obs.next();
          obs.complete();
        }
      });
    });
  }

  /**
   * This method publishes a message for a topic with optional options.
   * If an error occurs, it will throw.
   */
  public unsafePublish(topic: string, message: string | Buffer, options: IPublishOptions = {}): void {
    if (!this.client) {
      throw new Error('mqtt client not connected');
    }
    this.client.publish(topic, message, options, (error: Error | undefined) => {
      if (error) {
        throw (error);
      }
    });
  }

  private _handleOnConnect = (e: IOnConnectEvent) => {
    if (this.options.connectOnCreate === true) {
      Object.keys(this.observables).forEach((filterString: string) => {
        this.client.subscribe(filterString);
      });
    }
    this.state.next(MqttConnectionState.CONNECTED);
    this._onConnect.emit(e);
  }

  private _handleOnReconnect = () => {
    if (this.options.connectOnCreate === true) {
      Object.keys(this.observables).forEach((filterString: string) => {
        this.client.subscribe(filterString);
      });
    }
    this.state.next(MqttConnectionState.CONNECTING);
    this._onReconnect.emit();
  }

  private _handleOnClose = () => {
    this.state.next(MqttConnectionState.CLOSED);
    this._onClose.emit();
  }

  private _handleOnOffline = () => {
    this._onOffline.emit();
  }

  private _handleOnError = (e: IOnErrorEvent) => {
    this._onError.emit(e);
    console.error(e);
  }

  private _handleOnEnd = () => {
    this._onEnd.emit();
  }

  private _handleOnMessage = (topic: string, payload: Buffer, packet: Packet) => {
    this._onMessage.emit(packet);
    if (packet.cmd === 'publish') {
      this.messages.next(packet as any);
    }
  }

  private _handleOnPacketsend = (e: IOnPacketsendEvent) => {
    this._onPacketsend.emit(e);
  }

  private _handleOnPacketreceive = (e: IOnPacketreceiveEvent) => {
    this._onPacketreceive.emit(e);
  }

  private _generateClientId() {
    return 'client-' + Math.random().toString(36).substr(2, 19);
  }
}


export function provideMqtt(
  options: IMqttServiceOptions, client?: IMqttClient
): Provider[] {
  return [
    {
      provide: MqttServiceConfig,
      useValue: options
    },
    {
      provide: MqttClientService,
      useValue: client
    }
  ];
}
