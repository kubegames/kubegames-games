// The Vue build version to load with the `import` command
// (runtime-only or standalone) has been set in webpack.base.conf with an alias.
import Vue from 'vue'
import App from './App'
import { VueAxios } from './http/request'
import vuetify from '@/plugins/vuetify'
import Toast from "vue-toastification";
import "vue-toastification/dist/index.css";

const options = {
  //position: 'top-right',
  timeout: 2000,
  closeOnClick: true,
  pauseOnHover: true,
  draggable: true,
  draggablePercent: 0.6,
  showCloseButtonOnHover: false,
  hideProgressBar: true,
  closeButton: "button",
  icon: true
};

Vue.use(Toast, options);

Vue.use(VueAxios)

Vue.config.productionTip = false

/* eslint-disable no-new */
new Vue({
  el: '#app',
  vuetify,
  components: { App },
  template: '<App/>'
})
