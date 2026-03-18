db = db.getSiblingDB("privacy_db");

db.privacy_profiles.insertOne({
  public_key:
    "-----BEGIN RSA Public Key-----\nMIIBCgKCAQEAsEXAZrWESxKsjqsDk7LKjOOauiViBj8dNDw5Ssf+ikPWWsiEWdHo\nPN2lL7/5CL6+TZHhwj40uGwt9PTXs0Qgn5ZKKov572EusnMwTNTgEW1ZNkpX0aqW\nQyb2/47kfs3uz/zTNppxCeYLPWs8MKO3Xi1rknqD+LB96TQur4vJr+R4noRd85ya\npMF4xUerZzFxU1XhktwT7vTAdKqNN5UZkpiXCqtU7PbFz6el0Vik5pycdcjx+P6O\ny6J7G4J9BFpinAy6fzEbGB/++OALZosHTn7emoXkCBHqnxum2D1ue1VlX1wJXmWu\nf4wv3A/eKfud5ssErNvLK2JctKhVTZjVIQIDAQAB\n-----END RSA Public Key-----\n",
  personal_info: {
    name: "张三",
    id_card_num: "370102199001011234",
    credit_score: 750,
  },
  created_time: new Date(),
});
