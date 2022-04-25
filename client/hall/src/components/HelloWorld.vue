<template>
  <div data-app="true">
    
    <div class="text-center">
      <v-dialog
        v-model="dialog"
        width="500"
      >
        <v-card>
          <v-card-title class="text-h5 grey lighten-2">
            Welcome
          </v-card-title>

          <v-card-text>
            <v-form>
              <v-container>
                <v-row class="flex-column">
                  <v-form>
                    <v-col>
                      <v-text-field
                          v-model="phone"
                          :rules="phoneRules"
                          label="Your Phone Number"
                          required
                      ></v-text-field>
                      <v-text-field
                          v-model="password"
                          :rules="passRules"
                          type="password"
                          label="Password"
                          hint="At least 6 characters"
                          required
                      ></v-text-field>
                    </v-col>
                  </v-form>
                </v-row>
              </v-container>
            </v-form>
          </v-card-text>

          <v-divider></v-divider>

          <v-card-actions>
            <v-spacer></v-spacer>
            <v-btn
              color="primary"
              text
              :disabled="password.length === 0 || phone.length === 0"
              @click="login"
            >
              Login
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>
    </div>

    <ul v-for="(item,i) in games" v-bind:key="i">
      <li>
        <a @click='gotogame(item.path,item.gameID)'>
          {{ item.name }}
        </a>
      </li>
    </ul>
  </div>
</template>

<script>
import { Login, PlayerInfo } from '@/api/login';

export default {
  name: 'HelloWorld',
  data () {
    return {
      token:"",
      dialog: false,
      phone: '',
      phoneRules: [
        v => !!v || 'your phone number is required',
        v => /^1(3|4|5|6|7|8|9)\d{9}$/.test(v) || 'your phone number is incorrect',
      ],
      password: '',
      passRules: [
        v => !!v || 'password is required',
        v => v.length >= 6 || 'min 6 characters'
      ],
      games:[
        {name:"冰球突破",gameID:"90001",path:"bqtp"},
        {name:"财神到",gameID:"90002",path:"csd"},
        {name:"金瓶梅",gameID:"90003",path:"jinpingmei"},
        {name:"金玉满堂",gameID:"90004",path:"jinyumantang"},
        {name:"金鼠迎春",gameID:"90006",path:"jsyc"},
        {name:"金猪报福",gameID:"90007",path:"jzbf"},
        {name:"水果777",gameID:"90008",path:"sg777"},
        {name:"水果小玛丽",gameID:"90009",path:"sgxml"},
        {name:"水浒传",gameID:"90010",path:"shz"},
        {name:"五福临门",gameID:"90011",path:"wflm"},
        {name:"五龙争霸",gameID:"90012",path:"wlzb"},
        {name:"武士斩杀",gameID:"90013",path:"wszs"},
      ]
    }
  },
  mounted(){
    this.playerInfo()
  },
  methods:{
    gotogame(name,gameID){
      console.log(name,gameID);
      let url = window.location.protocol + "//" +window.location.hostname + "/web/"+name+"/index.html?token="+localStorage.getItem("token")+"&gameid="+gameID;
      console.log(url);
      window.open(url,"_self");
    },
    playerInfo(){
      PlayerInfo().then(response => {
        console.log("PlayerInfo",response)
        if (response.code === 200){
          this.dialog=false;
          if (response.data.status&&response.data.status.reson=="Success"){
            for(let j = 0; j < this.games.length; j++) {
              let game = this.games[j];
              if(game.gameID == response.data.status.gameID){
                this.gotogame(game.path,game.gameID);
              }
            }
          }
          return
        }
        this.dialog=true;
        localStorage.clear();
      }).catch(error => {
        console.log("login error",error)
        this.dialog=true;
        localStorage.clear();
      })
    },
    login(){
      Login({phone:this.phone, password:this.password}).then(response => {
        if (response.code === 200){
          this.token = response.data.token
          localStorage.setItem("token",this.token); 
          this.$toast.success("login success");
          this.playerInfo();
          this.dialog=false;
          return
        }
        this.$toast.error(response.msg);
      }).catch(error => {
        console.log("login error",error)
        this.$toast.error("login error");
      })
    },
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
h1, h2 {
  font-weight: normal;
}
ul {
  list-style-type: none;
  padding: 0;
}
li {
  display: inline-block;
  margin: 0 10px;
}
a {
  color: #42b983;
}
</style>
